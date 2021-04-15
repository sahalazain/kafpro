module github.com/sahalazain/kafpro

require (
	github.com/Shopify/sarama v1.28.0
	github.com/gofiber/fiber/v2 v2.4.1
	github.com/newrelic/newrelic-telemetry-sdk-go v0.5.1
	github.com/newrelic/opentelemetry-exporter-go v0.15.1
	github.com/psmarcin/fiber-opentelemetry v0.3.0
	github.com/sahalazain/go-common v0.0.0-20210412043808-764ed562529f
	github.com/segmentio/kafka-go v0.4.13
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.1
	go.opentelemetry.io/otel v0.16.0
	go.opentelemetry.io/otel/exporters/stdout v0.16.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.16.0
	go.opentelemetry.io/otel/sdk v0.16.0
	gocloud.dev v0.22.0
	gocloud.dev/pubsub/kafkapubsub v0.22.0
)

go 1.16
