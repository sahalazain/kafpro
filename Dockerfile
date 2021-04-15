FROM golang:1.15-alpine as builder

RUN apk update && apk upgrade && \
    apk --update add make

WORKDIR /app

COPY . .

RUN make build

FROM alpine
RUN apk --update --no-cache add ca-certificates && \
	addgroup -S kafpro && adduser -S -g kafpro kafpro
USER kafpro
COPY --from=builder /app/kafpro / 
CMD ["/kafpro", "serve"]
EXPOSE 8080