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
	subscriber *Subscriber
	clientID   ClientID

	// publish
	event Event
}

type Server struct {
	common.BaseService
	name string

	cmds chan cmd

	subscribers   map[ClientID]map[Topic]struct{}    // clientID -> topic -> empty struct
	subscriptions map[Topic]map[ClientID]*Subscriber // topic -> clientID -> subscriber

	// check if the subscriber has already been added before
	// subscribing or unsubscribing
	mtx sync.RWMutex
	wg  sync.WaitGroup
}

func NewServer(name string, logger log.Logger) *Server {
	server := &Server{
		name:        name,
		cmds:        make(chan cmd),
		subscribers: make(map[ClientID]map[Topic]struct{}),
	}
	server.BaseService = *common.NewBaseService(logger, name, server)
	return server
}

func (server *Server) OnStart() error {
	server.subscriptions = make(map[Topic]map[ClientID]*Subscriber)
	go server.loop()
	return nil
}

func (server *Server) OnStop() {
	server.cmds <- cmd{op: shutdown}
}

func (server *Server) HasSubscribed(clientID ClientID, topic Topic) bool {
	subs, ok := server.subscribers[clientID]
	if !ok {
		return ok
	}
	if len(topic) != 0 {
		_, ok = subs[topic]
	}
	return ok
}

func (server *Server) loop() {
	for cmd := range server.cmds {
		switch cmd.op {
		case unsub:
			if len(cmd.topic) != 0 {
				server.remove(cmd.clientID, cmd.topic)
			} else {
				server.removeClient(cmd.clientID)
			}
		case shutdown:
			server.removeAll()
			return
		case sub:
			// initialize subscription for this client per topic if needed
			if _, ok := server.subscriptions[cmd.topic]; !ok {
				server.subscriptions[cmd.topic] = make(map[ClientID]*Subscriber)
			}
			// create subscription
			server.subscriptions[cmd.topic][cmd.clientID] = cmd.subscriber
		case pub:
			server.push(cmd.event)
		}
	}
}

func (server *Server) push(event Event) {
	for _, sub := range server.subscriptions[event.GetTopic()] {
		sub.wg.Add(1)
		select {
		case sub.out <- event:
		default:
			// what should we do if channel is blocked? Remove the subscriber directly?
		}

		//go func(sub *Subscriber) {
		//	defer func() {
		//		if err := recover(); err != nil {
		//			publisher.Logger.Error("event handle err: ", err)
		//		}
		//		sub.wg.Done()
		//	}()
		//	handler, ok := sub.handlers[event.GetTopic()]
		//	if ok {
		//		handler(event)
		//	}
		//}(sub)
	}
	server.wg.Done()
}

func (server *Server) removeClient(clientID ClientID) {
	for topic, clientSubscriptions := range server.subscriptions {
		if _, ok := clientSubscriptions[clientID]; ok {
			server.remove(clientID, topic)
		}
	}
}

func (server *Server) removeAll() {
	for topic, clientSubscriptions := range server.subscriptions {
		for clientID := range clientSubscriptions {
			server.remove(clientID, topic)
		}
	}
}

func (server *Server) remove(clientID ClientID, topic Topic) {

	clientSubscriptions, ok := server.subscriptions[topic]
	if !ok {
		return
	}

	_, ok = clientSubscriptions[clientID]
	if !ok {
		return
	}
	// remove client from topic map.
	// if topic has no other clients subscribed, remove it.
	delete(server.subscriptions[topic], clientID)
	if len(server.subscriptions[topic]) == 0 {
		delete(server.subscriptions, topic)
	}
}

func (server *Server) Publish(e Event) {
	server.wg.Add(1)
	if !server.IsRunning() {
		return
	}

	select {
	case server.cmds <- cmd{op: pub, event: e}:
		return
	case <-server.Quit():
		return
	}
}
