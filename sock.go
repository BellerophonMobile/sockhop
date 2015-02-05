package sockhop

import (
	"github.com/BellerophonMobile/logberry"
	"github.com/gorilla/websocket"
	"net/http"
)

// Sock is a web socket connection.  It wraps a Gorilla websocket Conn
// and provides some basic call/response and other features.
type Sock struct {
	Socket *websocket.Conn
	Log *logberry.Task
}

var Upgrader websocket.Upgrader

// NewSock returns an unconnected Sock.  Parameter conf may be nil.
func NewSock(conf *SockConf) (*Sock, error) {

	var x Sock

	if conf == nil {
		conf = &SockConf{}
	}

	if conf.Context != nil {
		x.Log = conf.Context.Task("Web socket")
	} else {
		x.Log = logberry.Main.Task("Web socket")
	}
	
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
	
	return x, nil

}

func (x *Sock) Close() {
	x.Socket.Close()
}

func (x *Sock) ReportFailure(msg string) {	
}
