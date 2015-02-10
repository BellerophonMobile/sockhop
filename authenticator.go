package sockhop

import (
	"github.com/BellerophonMobile/logberry"
	"github.com/gorilla/websocket"
	"encoding/json"
)

type Authenticator interface {
	AuthenticateJWT(sock *Sock, jwt string) error
	AuthenticateUserPass(sock *Sock, user string, pass string) error
}

type authrequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	JWT      string `json:"jwt,omitempty"`
}

func (x *Sock) authenticate(conf *SockConf) error {

	if conf.Authenticator == nil {
		return nil
	}

	task := x.Log.Task("Authenticate")

	//-- First message must be authentication
	msgtype, message, err := x.Socket.ReadMessage()
	if err != nil {
		return task.Error(err)
	}

	if msgtype != websocket.TextMessage {
		return task.Failure("Expected credentials in text message")
	}

	var auth authrequest

	err = json.Unmarshal(message, &auth)
	if err != nil {
		return task.Error(err)
	}

	//-- Check the given credentials
	if auth.JWT != "" {
		err = conf.Authenticator.AuthenticateJWT(x, auth.JWT)
	} else if auth.Username != "" && auth.Password != "" {
		err = conf.Authenticator.AuthenticateUserPass(x, auth.Username, auth.Password)
	} else {
		err = logberry.NewError("Credentials expected but not given")
	}
	if err != nil {
		return task.Error(err)
	}

	return task.Success()
}

func (x *Sock) sendauthentication(conf *SockConf) error {

	if conf.JWT == "" && conf.User == "" && conf.Pass == "" {
		return nil
	}

	task := x.Log.Task("Authenticate")
	
	auth := &authrequest{JWT: conf.JWT, Username: conf.User, Password: conf.Pass}

	bytes, err := json.Marshal(auth)
	if err != nil {
		return task.WrapError("Could not marshal credentials", err)
	}

	err = x.Socket.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		return task.WrapError("Could not send credentials", err)
	}

	return task.Success()

}
