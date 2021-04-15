package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/sahalazain/go-common/config"
	"github.com/sahalazain/go-common/logger"
	kafka "github.com/segmentio/kafka-go"
)

type KafkaService struct {
	brokers []string
	subs    map[string]*kafka.Reader
	pub     *kafka.Writer
}

func NewKafkaService(conf config.Getter) (*KafkaService, error) {
	kb := conf.GetString("kafka_brokers")
	if kb == "" {
		return nil, errors.New("missing kafka_brokers config")
	}

	kbs := strings.Split(kb, ",")

	w := &kafka.Writer{
		Addr:     kafka.TCP(kbs[0]),
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaService{
		brokers: kbs,
		subs:    make(map[string]*kafka.Reader),
		pub:     w,
	}, nil
}

func (k *KafkaService) SendMessage(ctx context.Context, topic, key string, payload interface{}) error {

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: topic,
		Value: b,
	}

	if key != "" {
		msg.Key = []byte(key)
	}

	if err := k.pub.WriteMessages(ctx, msg); err != nil {
		return err
	}

	return nil
}

func (k *KafkaService) RetrieveMessage(ctx context.Context, topic, group string) (map[string]interface{}, error) {
	log := logger.GetLoggerContext(ctx, "service", "RetrieveMessage")

	sub, err := k.getSub(ctx, topic, group)
	if err != nil {
		return nil, err
	}

	//log.WithField("topic", topic).WithField("group", group).Debug("retrieve message")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()

	msg, err := sub.FetchMessage(ctx)
	if err != nil {
		log.WithError(err).Error("Error receiving message")
		return nil, err
	}

	var out map[string]interface{}
	if err := json.Unmarshal(msg.Value, &out); err != nil {
		return nil, err
	}

	if err := sub.CommitMessages(ctx, msg); err != nil {
		return nil, err
	}

	return out, nil
}

func (k *KafkaService) getSub(ctx context.Context, topic, group string) (*kafka.Reader, error) {
	if s, ok := k.subs[topic]; ok {
		return s, nil
	}

	kcf := kafka.ReaderConfig{
		Brokers:        k.brokers,
		Topic:          topic,
		MinBytes:       10,   // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Duration(10),
		QueueCapacity:  1,
		GroupID:        group,
	}

	sub := kafka.NewReader(kcf)
	k.subs[topic] = sub
	return sub, nil
}
