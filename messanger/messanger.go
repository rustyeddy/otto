package messanger

// Subscriber is an interface that defines a struct needs to have the
// Callback(topic string, data []byte) function defined.
type MsgHandler func(msg *Msg)

type Messanger interface {
	ID() string
	Subscribe(topic string, handler MsgHandler) error
	SetTopic(topic string)
	PubMsg(msg *Msg)
	PubData(data any)
	Error() error
	Close()
}
