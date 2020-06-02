package pubsub

import "sync"

type ClientID string

type subscriber struct {
	clientID ClientID
	pub      *Publisher
	handlers map[Topic]Handler
	wg       *sync.WaitGroup
}

func (publisher *Publisher) NewSubscriber(clientID ClientID) (*subscriber, error) {
	publisher.mtx.Lock()
	defer publisher.mtx.Unlock()
	_, ok := publisher.subscribers[clientID]
	if ok {
		return nil, ErrDuplicateClientID
	}
	sub := &subscriber{
		clientID: clientID,
		pub:      publisher,
		handlers: make(map[Topic]Handler),
		wg:       &sync.WaitGroup{},
	}
	publisher.subscribers[clientID] = make(map[Topic]struct{})
	return sub, nil
}

func (s *subscriber) Subscribe(topic Topic, handler Handler) error {
	if handler == nil {
		return ErrNilHandler
	}
	s.pub.mtx.RLock()
	subscribers, ok := s.pub.subscribers[s.clientID]
	if ok {
		_, ok = subscribers[topic]
	}
	s.pub.mtx.RUnlock()
	if ok {
		return ErrAlreadySubscribed
	}

	s.handlers[topic] = handler

	select {
	case s.pub.cmds <- cmd{op: sub, topic: topic, subscriber: s, clientID: s.clientID}:
		s.pub.mtx.Lock()
		if _, ok := s.pub.subscribers[s.clientID]; !ok {
			s.pub.subscribers[s.clientID] = make(map[Topic]struct{})
		}
		s.pub.subscribers[s.clientID][topic] = struct{}{}
		s.pub.mtx.Unlock()
		return nil
	case <-s.pub.Quit():
		return nil
	}
}

func (s *subscriber) Unsubscribe(topic Topic) error {
	s.pub.mtx.RLock()
	subscribers, ok := s.pub.subscribers[s.clientID]
	if ok {
		_, ok = subscribers[topic]
	}
	s.pub.mtx.RUnlock()
	if !ok {
		return ErrSubscriptionNotFound
	}
	select {
	case s.pub.cmds <- cmd{op: unsub, clientID: s.clientID, topic: topic}:
		s.pub.mtx.Lock()
		delete(s.pub.subscribers[s.clientID], topic)
		s.pub.mtx.Unlock()
		return nil
	case <-s.pub.Quit():
		return nil
	}
}

func (s *subscriber) UnsubscribeAll() error {
	s.pub.mtx.RLock()
	_, ok := s.pub.subscribers[s.clientID]
	s.pub.mtx.RUnlock()
	if !ok {
		return ErrSubscriptionNotFound
	}
	select {
	case s.pub.cmds <- cmd{op: unsub, clientID: s.clientID}:
		s.pub.mtx.Lock()
		delete(s.pub.subscribers, s.clientID)
		s.pub.mtx.RUnlock()
		return nil
	case <-s.pub.Quit():
		return nil
	}
}

func (s *subscriber) Wait() {
	s.wg.Wait()
}
