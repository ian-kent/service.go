package producer

import "github.com/Shopify/sarama"

// Producer ...
type Producer interface {
	Send(Message) (partition int32, offset int64, err error)
}

// Message ...
type Message interface {
	Topic() string
	Key() sarama.Encoder
	Value() sarama.Encoder
}

type saramaMessage struct {
	key, value sarama.Encoder
	topic      string
}

func (sm saramaMessage) Key() sarama.Encoder   { return sm.key }
func (sm saramaMessage) Value() sarama.Encoder { return sm.value }
func (sm saramaMessage) Topic() string         { return sm.topic }

// NewStringMessage returns a new Message
func NewStringMessage(topic, key, value string) Message {
	return saramaMessage{sarama.StringEncoder(key), sarama.StringEncoder(value), topic}
}

// NewByteMessage returns a new Message
func NewByteMessage(topic, key string, value []byte) Message {
	return saramaMessage{sarama.StringEncoder(key), sarama.ByteEncoder(value), topic}
}

type kafkaProducer struct {
	producer sarama.SyncProducer

	Config
}

// New ...
func New(config Config) (Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 10

	// TODO: TLS config, see https://github.com/Shopify/sarama/blob/master/examples/http_server/http_server.go

	producer, err := sarama.NewSyncProducer(config.KafkaBrokers(), cfg)
	if err != nil {
		return nil, err
	}

	return &kafkaProducer{
		Config:   config,
		producer: producer,
	}, nil
}

func (kc *kafkaProducer) Send(msg Message) (partition int32, offset int64, err error) {
	return kc.producer.SendMessage(&sarama.ProducerMessage{
		Key:   msg.Key(),
		Value: msg.Value(),
		Topic: msg.Topic(),
	})
}
