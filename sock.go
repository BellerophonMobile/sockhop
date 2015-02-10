package sockhop

import (
	"sync/atomic"
	"github.com/BellerophonMobile/logberry"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
)

type Hook func(*Message) error

// Sock is a web socket connection.  It wraps a Gorilla websocket Conn
// and provides some basic call/response and other features.
type Sock struct {
	Socket *websocket.Conn

	Log *logberry.Task
	
	TextHooks map[string]Hook

	BinaryHandler func([]byte) error
	ErrorHandler func(error) error
	
	doneloop chan bool
	messages uint64

	Live bool
}

var Upgrader websocket.Upgrader

// NewSock returns an unconnected Sock.  Parameter conf may be nil.
func NewSock(conf *SockConf) (*Sock, error) {

	var x Sock

	if conf == nil {
		conf = &SockConf{}
	}

	if conf.Context == nil {
		x.Log = logberry.Main.Task("Websocket")
	} else {
		x.Log = conf.Context.Task("Websocket")
	}
	
	x.ErrorHandler = conf.ErrorHandler
	x.BinaryHandler = conf.BinaryHandler
	
	x.TextHooks = make(map[string]Hook)
	
	x.doneloop = make(chan bool, 1)
	
	return &x, nil
}


// Upgrade converts an HTTP connection into a web socket.  An active
// Sock is returned, or an error if there is a problem.
func Upgrade(w http.ResponseWriter, r *http.Request, conf *SockConf) (*Sock, error) {

	if conf == nil {
		conf = &SockConf{}
	}

	var err error

	x, err := NewSock(conf)
	if err != nil {
		return nil, err
	}
	
	//-- Upgrade the connection to a websocket
	x.Socket, err = Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, x.Log.WrapError("Could not upgrade to websocket", err)
	}
	
	if conf.MaxMessageSize != 0 {	
		x.Socket.SetReadLimit(conf.MaxMessageSize)
	}

	err = x.authenticate(conf)
	if err != nil {
		return nil, err
	}

	x.Live = true

	x.Log.Info("Connected")
	
	return x, nil

}

func Dial(url string, conf *SockConf) (*Sock, error) {

	if conf == nil {
		conf = &SockConf{}
	}

	var err error

	x, err := NewSock(conf)
	if err != nil {
		return nil, err
	}

	dialer := &websocket.Dialer{
		TLSClientConfig: conf.TLSConfig,
		ReadBufferSize:  conf.ReadBufferSize,
		WriteBufferSize: conf.WriteBufferSize,
	}

	task := x.Log.Task("Connect").Service(url).Time()

	x.Socket, _, err = dialer.Dial(url, nil)
	if err != nil {
		return nil, task.Error(err)
	}

	if conf.MaxMessageSize != 0 {	
		x.Socket.SetReadLimit(conf.MaxMessageSize)
	}

	err = x.sendauthentication(conf)
	if err != nil {
		return nil, task.Error(err)
	}
	
	x.Live = true
	
	task.Success()

	return x, nil
	
}

func (x *Sock) AddHook(code string, hook Hook) {
	x.TextHooks[code] = hook
}

func (x *Sock) Close() {
	x.Live = false
	x.doneloop <- true
	x.Socket.Close()
}


type socketproduct struct {
	msgtype int
	data []byte
	err error
}


func (x *Sock) readmessage(ch chan socketproduct) {
	var sp socketproduct
	sp.msgtype, sp.data, sp.err = x.Socket.ReadMessage()
	ch <- sp
}

func (x *Sock) ReadMessage() (*Message,error) {
	msgtype,data,err := x.Socket.ReadMessage()
	if err != nil { return nil, err}

	if msgtype != websocket.TextMessage {
		return nil, x.Log.Failure("Expected text message")
	}

	var message	Message
	message.Sock = x

	err = json.Unmarshal(data, &message)
	if err != nil {
		return nil,err
	}

	return &message,nil

}

func (x *Sock) Loop() error {

	readmsg := make(chan socketproduct, 1)

	var done bool
	var err error
	
	for !done && err == nil {

		go x.readmessage(readmsg)

		select {
		case msg := <- readmsg:

			if msg.err != nil {
				err = x.processError(msg.err)
				continue
			}
			
			switch msg.msgtype {
			case websocket.TextMessage:
				err = x.processText(msg.data)

			case websocket.BinaryMessage:
				err = x.processBinary(msg.data)
			}
			
		case <- x.doneloop:
			done = true
		}
		
	}

	return err

}

func (x *Sock) processError(err error) error {
	if x.ErrorHandler != nil {
		return x.ErrorHandler(err)
	}

	return err
}

func (x *Sock) processText(data []byte) error {

	var message	Message
	message.Sock = x

	err := json.Unmarshal(data, message)
	if err != nil {
		return err
	}

	/*
	if message.inResponseTo != 0 {
		// Look up what it was in response to, and handle it
	}
	*/

	fn,ok := x.TextHooks[message.Code]
	if !ok {
		return x.Log.Failure("Received unexpected text message for code ",
			message.Code)
	}

	return fn(&message)

}

func (x *Sock) processBinary(data []byte) error {

	if x.BinaryHandler == nil {
		return x.Log.Failure("Received unexpected binary message")
	}
	
	return x.BinaryHandler(data)
	
}

func (x *Sock) SendFault(msg string) error {
	
	return x.SendMessage("ERROR", []byte(msg))
	
}

func (x *Sock) SendMessage(code string, data interface{}) error {

	var err error
	
	var message Message
	message.Code = code
	message.ID = atomic.AddUint64(&x.messages, 1)
	bytes,err := json.Marshal(data)
	if err != nil {
		return x.Log.WrapError("Could not marshal data", err)
	}
	message.Data = string(bytes)	

	bytes, err = json.Marshal(message)
	if err != nil {
		return x.Log.WrapError("Could not marshal message", err)
	}

	err = x.Socket.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		return x.Log.WrapError("Could not write command", err)
	}

	return nil

}

func (x *Sock) SendText(data []byte) error {
	if !x.Live {
		return logberry.NewError("Sock is not live")
	}

	return x.Socket.WriteMessage(websocket.TextMessage, data)
}

func (x *Sock) SendBinary(data []byte) error {
	if !x.Live {
		return logberry.NewError("Sock is not live")
	}

	return x.Socket.WriteMessage(websocket.BinaryMessage, data)
}
