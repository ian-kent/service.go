package consumer

import (
	"os"
	"os/signal"

	golog "log"

	"github.com/Shopify/sarama"
	"github.com/ian-kent/service.go/log"
	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"
)

func init() {
	sarama.Logger = golog.New(os.Stdout, "[Sarama] ", golog.LstdFlags)
}

// Consumer ...
type Consumer interface {
	Start() chan Message
	Commit(to Message) error
}

// Message ...
type Message interface {
	Key() []byte
	Value() []byte
	Partition() int32
	Offset() int64
}

type saramaMessage struct {
	*sarama.ConsumerMessage
}

func (sm saramaMessage) Key() []byte      { return sm.ConsumerMessage.Key }
func (sm saramaMessage) Value() []byte    { return sm.ConsumerMessage.Value }
func (sm saramaMessage) Partition() int32 { return sm.ConsumerMessage.Partition }
func (sm saramaMessage) Offset() int64    { return sm.ConsumerMessage.Offset }

type kafkaConsumer struct {
	consumerGroup *consumergroup.ConsumerGroup
	sigChan       chan os.Signal

	Config
}

// New ...
func New(config Config) Consumer {
	return &kafkaConsumer{
		Config: config,
	}
}

func (kc *kafkaConsumer) Commit(to Message) error {
	return kc.consumerGroup.CommitUpto(to.(saramaMessage).ConsumerMessage)
}

func (kc *kafkaConsumer) Start() chan Message {
	kc.sigChan = make(chan os.Signal, 1)
	msgChan := make(chan Message, 1)

	signal.Notify(kc.sigChan, os.Interrupt)
	go func() {
		<-kc.sigChan
		kc.consumerGroup.Close()
		close(msgChan)
	}()

	cfg := consumergroup.NewConfig()

	cfg.Offsets.Initial = kc.Config.InitialOffset()
	cfg.Offsets.ProcessingTimeout = kc.Config.ProcessingTimeout()

	var zookeeperNodes []string
	url := kc.Config.ZookeeperURL()
	if chroot := kc.Config.ZookeeperChroot(); len(chroot) > 0 {
		url += "/" + chroot
	}
	zookeeperNodes, cfg.Zookeeper.Chroot = kazoo.ParseConnectionString(url)

	cg, err := consumergroup.JoinConsumerGroup(
		kc.Config.ConsumerGroup(),
		kc.Config.Topics(),
		zookeeperNodes,
		cfg,
	)

	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	kc.consumerGroup = cg

	go func() {
		for err := range cg.Errors() {
			log.Error(err, nil)
		}
	}()

	go func() {
		log.Debug("waiting for messages", nil)
		for m := range cg.Messages() {
			log.Debug("message", log.Data{"msg": m})
			msgChan <- saramaMessage{m}
		}
	}()

	return msgChan
}
