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

	var authrequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		JWT      string `json:"jwt"`
	}

	err = json.Unmarshal(message, &authrequest)
	if err != nil {
		return task.Error(err)
	}

	
	//-- Check the given credentials
	if authrequest.JWT != "" {
		err = conf.Authenticator.AuthenticateJWT(x, authrequest.JWT)
	} else if authrequest.Username != "" && authrequest.Password != "" {
		err = conf.Authenticator.AuthenticateUserPass(x, authrequest.Username, authrequest.Password)
	} else {
		err = logberry.NewError("Credentials expected but not given")
	}
	if err != nil {
		return task.Error(err)
	}

	return task.Success()
}
