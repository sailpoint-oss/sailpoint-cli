VERSION = 0.1.0

clean:
	go clean ./...

mocks:
	# Ref: https://github.com/golang/mock
	mockgen -source=client/client.go -destination=mocks/client.go -package=mocks

test:
	docker build -t cli .
	docker run --rm cli go test -v -count=1 ./...

install:
	go build -o /usr/local/bin/sail -ldflags="-X 'github.com/sailpoint-oss/sailpoint-cli/cmd/root.version=$(VERSION)'"

.PHONY: clean mocks test install .docker/login .docker/build .docker/push
