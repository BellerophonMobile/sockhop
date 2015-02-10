package sockhop

import (
	"github.com/BellerophonMobile/logberry"
	"crypto/tls"
)

// SockConf encapsulates optional parameters to configure a Sock.
type SockConf struct {
	MaxMessageSize int64
	ReadBufferSize int
	WriteBufferSize int

	Authenticator Authenticator

	JWT string
	User string
	Pass string

	BinaryHandler func(data []byte) error
	
	ErrorHandler func(error) error

	TLSConfig *tls.Config
	
	Context logberry.Context
}
