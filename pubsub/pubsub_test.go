package pubsub

import (
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/pubsub"
	"testing"
	"time"
)

const blockT = Topic("block")

type BlockCompleteEvent struct {
	txNum int
}

func (bc BlockCompleteEvent) GetTopic() Topic {
	return blockT
}

func TestSubscribe(t *testing.T) {

	pub := startPublisher(t, "test_pubsub")

	sub, err := pub.NewSubscriber("test_client")
	require.Nil(t, err)

	_, err = pub.NewSubscriber("test_client")
	require.Equal(t, ErrDuplicateClientID, err)

	var getTxNum int
	err = sub.Subscribe(blockT, func(event Event) {
		switch event.(type) {
		case BlockCompleteEvent:
			time.Sleep(time.Second)
			bc := event.(BlockCompleteEvent)
			getTxNum = bc.txNum
		}
	})
	require.Nil(t, err)
	err = sub.Subscribe(blockT, func(event Event) {})
	require.Equal(t, pubsub.ErrAlreadySubscribed, err)

	pub.Publish(BlockCompleteEvent{txNum: 100})
	require.NotEqual(t, 100, getTxNum)
	sub.Wait()
	require.Equal(t, 100, getTxNum)

}

func TestUnsubscribe(t *testing.T) {
	pub := startPublisher(t, "test_pubsub")

	clientId := ClientID("test_client")
	sub, err := pub.NewSubscriber(clientId)
	require.Nil(t, err)

	err = sub.Subscribe(blockT, func(event Event) {})
	require.Nil(t, err)

	require.True(t, pub.HasSubscribed(clientId, blockT))

	err = sub.Unsubscribe(blockT)
	require.Nil(t, err)

	require.False(t, pub.HasSubscribed(clientId, blockT))
}

func startPublisher(t *testing.T, name string) *Publisher {
	pub := NewPublisher(name, nil)
	err := pub.Start()
	require.Nil(t, err)
	return pub
}
