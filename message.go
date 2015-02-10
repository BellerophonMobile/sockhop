package sockhop

type Message struct {

	Code string
	ID uint64
	InResponseTo uint64
	Data string

	Sock *Sock `json:"-"`
}
