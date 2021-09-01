package connections

import (
	"fmt"
	"time"

	"github.com/enorith/framework/queue/std"
	"github.com/nsqio/go-nsq"
	"github.com/vmihailenco/msgpack/v5"
)

type Config struct {
	Lookupd, Nsqd, Topic, Channel string
	UsingLookup, valid            bool
}

type Nsq struct {
	config    map[string]interface{}
	configVal Config
	consumer  *nsq.Consumer
	producer  *nsq.Producer
}

func (n *Nsq) Consume(concurrency int) (err error) {
	c := nsq.NewConfig()
	config := n.configVal
	n.consumer, err = nsq.NewConsumer(config.Topic, config.Channel, c)
	if err != nil {
		return err
	}

	// consumer.AddHandler(&myMessageHandler{})

	if config.UsingLookup {
		return n.consumer.ConnectToNSQLookupd(config.Lookupd)
	}

	return n.consumer.ConnectToNSQD(config.Nsqd)
}

func (n *Nsq) Stop() error {
	if n.consumer != nil {
		n.consumer.Stop()
	}
	return nil
}

func (n *Nsq) Dispatch(payload interface{}, delay time.Duration) (err error) {
	c := nsq.NewConfig()
	config := n.configVal
	n.producer, err = nsq.NewProducer(config.Nsqd, c)
	if err != nil {
		return
	}
	n.producer.SetLogger(nil, 0)
	var messageBody []byte
	messageBody, err = msgpack.Marshal(payload)

	if err != nil {
		return
	}
	if delay == 0 {
		err = n.producer.Publish(config.Topic, messageBody)
	} else {
		err = n.producer.DeferredPublish(config.Topic, delay, messageBody)
	}

	return
}

func (n *Nsq) parseConfig() (c Config, e error) {
	if n.configVal.valid {
		return n.configVal, nil
	}

	c.Topic = "default"

	var is bool

	if lkupd, ok := n.config["nsqlookupd"]; ok {
		if c.Lookupd, is = lkupd.(string); !is {
			return c, fmt.Errorf("nsq config: nsqlookupd should be string, %T giving", lkupd)
		}
		c.UsingLookup = true
	}

	if nsqd, ok := n.config["nsqd"]; ok {
		if c.Nsqd, is = nsqd.(string); !is {
			return c, fmt.Errorf("nsq config: nsqd should be string, %T giving", nsqd)
		}
	}

	if topic, ok := n.config["topic"]; ok {
		if c.Topic, is = topic.(string); !is {
			return c, fmt.Errorf("nsq config: topic should be string, %T giving", topic)
		}
	}

	if channel, ok := n.config["channel"]; ok {
		if c.Channel, is = channel.(string); !is {
			return c, fmt.Errorf("nsq config: channel should be string, %T giving", channel)
		}
	}
	c.valid = true
	n.configVal = c

	return
}

func NewNsq(config map[string]interface{}) *Nsq {
	nsq := &Nsq{config: config}
	nsq.parseConfig()
	return nsq
}

type JobHandler struct {
}

func (JobHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		return nil
	}
	var job std.Job

	msgpack.Unmarshal(m.Body, &job)
	return nil
}
