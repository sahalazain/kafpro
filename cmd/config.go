package cmd

//DefaultConfig default config
var DefaultConfig = map[string]interface{}{
	"port":            "8080",
	"name":            "kafpro",
	"log_level":       "DEBUG",
	"log_format":      "json",
	"tracer":          "no-op",
	"newrelic_apikey": "",
	"jaeger_url":      "http://localhost:14268/api/traces",
	"kafka_brokers":   "localhost:9092",
	"service_type":    "kafka",
	"read_timeout":    10,
}
