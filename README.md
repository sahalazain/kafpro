# kafpro
Kafka REST API Proxy


# Build
```
make dependencies
make build
```
```
make dev
```
# Build Docker
```make docker```

# Configuration (environment variable)
``` 
KAFKA_BROKERS=localhost:9092
READ_TIMEOUT=20
```
## Run
```env KAFKA_BROKERS=localhost:9092 ./kafpro serve```

# Usage

## Retrieve Message from Topic
```GET /topic/{topic_name}/{consumer_group}```

## Bulk Retrieve
```GET /topic/{topic_name}/{consumer_group}?max={number_of_message}```

## Send Message (JSON)
```POST /topic/{topic_name} @body```
