BINARY=kafpro

dev: 
	@go mod tidy
	@go mod vendor
	@go build


build:
	env GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY} -mod=vendor -a -installsuffix cgo -ldflags '-w'

dependencies:
	@echo "> Installing the server dependencies ..."
	@go mod download
	@go mod vendor

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

run:
	./kafpro serve

run-dev: dev run

docker: 
	@echo "> Build Docker image"
	@docker build -t sahalazain/kafpro . 


.PHONY: clean build docs run test run-dev docker
