package pubsub

type Topic string

const (
	CrossTransferTopic = Topic("cross-transfer")
)

type Event interface {
	GetTopic() Topic
}

type Handler func(Event)

type CrossReceiver struct {
	Addr   string
	Amount int64
}

type CrossTransferEvent struct {
	TxHash     string
	ChainId    string
	Type       string
	RelayerFee int64
	From       string
	Denom      string
	Contract   string
	Decimals   int
	To         []CrossReceiver
}

func (event CrossTransferEvent) GetTopic() Topic {
	return CrossTransferTopic
}
