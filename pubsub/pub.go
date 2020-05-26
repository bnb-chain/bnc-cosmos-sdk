package pubsub

import (
	"errors"
	"sync"

	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	// ErrDuplicateSubscriber is returned when a client tries to subscribe
	// with an existing client ID.
	ErrDuplicateClientID = errors.New("clientID is exist")

	// ErrAlreadySubscribed is returned when a client tries to subscribe twice or
	// more using the same topic.
	ErrAlreadySubscribed = errors.New("already subscribed")

	// ErrSubscriptionNotFound is returned when a client tries to unsubscribe
	// from not existing subscription.
	ErrSubscriptionNotFound = errors.New("subscription not found")

	ErrNilHandler = errors.New("handler is nil")
)

type operation int

const (
	sub operation = iota
	pub
	unsub
	shutdown
)

type cmd struct {
	op operation

	// subscribe, unsubscribe
	topic      Topic
	subscriber *subscriber
	clientID   ClientID

	// publish
	event Event
}

type Publisher struct {
	common.BaseService
	name string

	cmds chan cmd

	subscribers   map[ClientID]map[Topic]struct{}    // clientID -> topic -> empty struct
	subscriptions map[Topic]map[ClientID]*subscriber // topic -> clientID -> subscriber

	mtx sync.RWMutex
}

func NewPublisher(name string, logger log.Logger) *Publisher {
	publisher := &Publisher{
		name:        name,
		cmds:        make(chan cmd),
		subscribers: make(map[ClientID]map[Topic]struct{}),
	}
	publisher.BaseService = *common.NewBaseService(logger, name, publisher)
	return publisher
}

func (publisher *Publisher) OnStart() error {
	publisher.subscriptions = make(map[Topic]map[ClientID]*subscriber)
	go publisher.loop()
	return nil
}

func (publisher *Publisher) OnStop() {
	publisher.cmds <- cmd{op: shutdown}
}

func (publisher *Publisher) HasSubscribed(clientID ClientID, topic Topic) bool {
	subs, ok := publisher.subscribers[clientID]
	if !ok {
		return ok
	}
	if len(topic) != 0 {
		_, ok = subs[topic]
	}
	return ok
}

func (publisher *Publisher) loop() {
loop:
	for cmd := range publisher.cmds {
		switch cmd.op {
		case unsub:
			if len(cmd.topic) != 0 {
				publisher.remove(cmd.clientID, cmd.topic)
			} else {
				publisher.removeClient(cmd.clientID)
			}
		case shutdown:
			publisher.removeAll()
			break loop
		case sub:
			// initialize subscription for this client per topic if needed
			if _, ok := publisher.subscriptions[cmd.topic]; !ok {
				publisher.subscriptions[cmd.topic] = make(map[ClientID]*subscriber)
			}
			// create subscription
			publisher.subscriptions[cmd.topic][cmd.clientID] = cmd.subscriber
		case pub:
			publisher.push(cmd.event)
		}
	}
}

func (publisher *Publisher) push(event Event) {
	for _, subscriber := range publisher.subscriptions[event.GetTopic()] {
		subscriber.wg.Add(1)
		go func() {
			subscriber.handlers[event.GetTopic()](event)
			subscriber.wg.Done()
		}()
	}

}

func (publisher *Publisher) removeClient(clientID ClientID) {
	for topic, clientSubscriptions := range publisher.subscriptions {
		if _, ok := clientSubscriptions[clientID]; ok {
			publisher.remove(clientID, topic)
		}
	}
}

func (publisher *Publisher) removeAll() {
	for topic, clientSubscriptions := range publisher.subscriptions {
		for clientID := range clientSubscriptions {
			publisher.remove(clientID, topic)
		}
	}
}

func (publisher *Publisher) remove(clientID ClientID, topic Topic) {

	clientSubscriptions, ok := publisher.subscriptions[topic]
	if !ok {
		return
	}

	_, ok = clientSubscriptions[clientID]
	if !ok {
		return
	}
	// remove client from topic map.
	// if topic has no other clients subscribed, remove it.
	delete(publisher.subscriptions[topic], clientID)
	if len(publisher.subscriptions[topic]) == 0 {
		delete(publisher.subscriptions, topic)
	}
}

func (publisher *Publisher) Publish(e Event) {
	if !publisher.IsRunning() {
		return
	}
	select {
	case publisher.cmds <- cmd{op: pub, event: e}:
		return
	case <-publisher.Quit():
		return
	}
}
