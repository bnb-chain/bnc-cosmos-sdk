package pubsub

type Topic string

type Event interface {
	GetTopic() Topic
	//FromTx() bool
}

type Handler func(Event)
