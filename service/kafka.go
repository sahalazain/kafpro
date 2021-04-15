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

const defaultTimeout = 10

type KafkaService struct {
	brokers []string
	subs    map[string]*kafka.Reader
	pub     *kafka.Writer
	timeout int
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
	to := conf.GetInt("read_timeout")
	if to == 0 {
		to = defaultTimeout
	}

	return &KafkaService{
		brokers: kbs,
		subs:    make(map[string]*kafka.Reader),
		pub:     w,
		timeout: to,
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

func (k *KafkaService) BulkRetrieve(ctx context.Context, topic, group string, max int) ([]map[string]interface{}, error) {
	log := logger.GetLoggerContext(ctx, "service", "BulkRetrieve")

	sub, err := k.getSub(ctx, topic, group)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(k.timeout)*time.Second)
	defer cancel()

	outs := make([]map[string]interface{}, 0)

	for i := 0; i < max; i++ {
		msg, err := sub.FetchMessage(ctx)
		if err != nil {
			if err == context.DeadlineExceeded && len(outs) > 0 {
				return outs, nil
			}
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
		outs = append(outs, out)
	}

	return outs, nil
}

func (k *KafkaService) RetrieveMessage(ctx context.Context, topic, group string) (map[string]interface{}, error) {
	log := logger.GetLoggerContext(ctx, "service", "RetrieveMessage")

	sub, err := k.getSub(ctx, topic, group)
	if err != nil {
		return nil, err
	}

	//log.WithField("topic", topic).WithField("group", group).Debug("retrieve message")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(k.timeout)*time.Second)
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
	if s, ok := k.subs[topic+group]; ok {
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
	k.subs[topic+group] = sub
	return sub, nil
}
