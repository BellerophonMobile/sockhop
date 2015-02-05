package sockhop

import (
	"github.com/BellerophonMobile/logberry"
)

// SockConf encapsulates optional parameters to configure a Sock.
type SockConf struct {
	Context logberry.Context

	MaxMessageSize int64

	Authenticator Authenticator
}
