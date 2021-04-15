package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/sahalazain/go-common/config"
	"github.com/sahalazain/go-common/logger"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/kafkapubsub"
)

const keyName = "key"

type App interface {
	SendMessage(ctx context.Context, topic, key string, payload interface{}) error
	RetrieveMessage(ctx context.Context, topic, group string) (map[string]interface{}, error)
}

func GetService(conf config.Getter) (App, error) {
	st := conf.GetString("service_type")
	switch strings.ToLower(st) {
	case "pubsub":
		return NewPubsubService(conf)
	case "kafka":
		return NewKafkaService(conf)
	default:
		return nil, errors.New("unsupported service type")
	}
}

type PubsubService struct {
	brokers []string
	pubs    map[string]*pubsub.Topic
	subs    map[string]*pubsub.Subscription
	Config  config.Getter
}

func NewPubsubService(conf config.Getter) (*PubsubService, error) {
	kb := conf.GetString("kafka_brokers")
	if kb == "" {
		return nil, errors.New("missing kafka_brokers config")
	}

	os.Setenv("KAFKA_BROKERS", kb)
	return &PubsubService{
		brokers: strings.Split(kb, ","),
		Config:  conf,
		pubs:    make(map[string]*pubsub.Topic),
		subs:    make(map[string]*pubsub.Subscription),
	}, nil
}

func (a *PubsubService) SendMessage(ctx context.Context, topic, key string, payload interface{}) error {
	pub, err := a.getPub(ctx, topic)
	if err != nil {
		return err
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := &pubsub.Message{Body: b}
	if key != "" {
		msg.Metadata[keyName] = key
	}

	if err := pub.Send(ctx, msg); err != nil {
		return err
	}

	return nil
}

func (a *PubsubService) RetrieveMessage(ctx context.Context, topic, group string) (map[string]interface{}, error) {
	log := logger.GetLoggerContext(ctx, "service", "RetrieveMessage")

	sub, err := a.getSub(ctx, topic, group)
	if err != nil {
		return nil, err
	}

	//defer sub.Shutdown(ctx)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	msg, err := sub.Receive(ctx)
	if err != nil {
		log.WithError(err).Error("error retrieving message")
		return nil, err
	}

	var out map[string]interface{}
	if err := json.Unmarshal(msg.Body, &out); err != nil {
		log.WithError(err).Error("error unmarshal")
		return nil, err
	}

	msg.Ack()

	return out, nil

}

func (a *PubsubService) getPub(ctx context.Context, topic string) (*pubsub.Topic, error) {

	if t, ok := a.pubs[topic]; ok {
		return t, nil
	}

	cfg := kafkapubsub.MinimalConfig()
	cfg.Consumer.Fetch.Min = 1
	pub, err := kafkapubsub.OpenTopic(a.brokers, cfg, topic, &kafkapubsub.TopicOptions{KeyName: keyName})

	if err != nil {
		return nil, err
	}
	a.pubs[topic] = pub
	return pub, nil
}

func (a *PubsubService) getSub(ctx context.Context, topic, group string) (*pubsub.Subscription, error) {

	if s, ok := a.subs[topic]; ok {
		return s, nil
	}

	cfg := kafkapubsub.MinimalConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	sub, err := kafkapubsub.OpenSubscription(a.brokers, cfg, group, []string{topic}, &kafkapubsub.SubscriptionOptions{KeyName: keyName})
	if err != nil {
		return nil, err
	}

	a.subs[topic] = sub
	return sub, nil
}
