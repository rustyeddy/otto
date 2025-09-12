package messanger

// Subscriber is an interface that defines a struct needs to have the
// Callback(topic string, data []byte) function defined.
type MsgHandler func(msg *Msg)

type Messanger interface {
	ID() string
	Subscribe(topic string, handler MsgHandler) error
	SetTopic(topic string)
	Topic() string
	PubMsg(msg *Msg)
	PubData(data any)
	Error() error
	Close()
}

// MessangerBase
type MessangerBase struct {
	id    string
	topic []string
	subs  map[string][]MsgHandler

	Published int
}

func NewMessangerBase(id string, topic ...string) MessangerBase {
	return MessangerBase{
		id:    id,
		topic: topic,
		subs:  make(map[string][]MsgHandler),
	}
}

func (mb *MessangerBase) ID() string {
	return mb.id
}

func (mb *MessangerBase) Topic() string {
	return mb.topic[0]
}

func (mb *MessangerBase) SetTopic(topic string) {
	mb.topic = append(mb.topic, topic)
}
