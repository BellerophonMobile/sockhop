package sockhop

import (
	"github.com/BellerophonMobile/logberry"
)

type Message struct {

	Code string
	ID uint64
	InResponseTo uint64
	Data string

	Sock *Sock `json:"-"`
}


func (x *Message) Reply(code string, data interface{}) error {

	if x.Sock == nil {
		return logberry.NewError("Inactive message")
	}
	
	message, err := x.Sock.newmessage(code, data)
	if err != nil {
		return nil
	}

	message.InResponseTo = x.ID

	return x.Sock.sendmessage(message)
	
}
