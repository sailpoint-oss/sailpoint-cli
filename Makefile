clean:
	go clean ./...

mocks:
	# Ref: https://github.com/golang/mock
	mockgen -source=client/client.go -destination=mocks/client.go -package=mocks

test:
	go test -v -count=1 ./...

install:
	go build -o /usr/local/bin/sail

.PHONY: clean mocks test install .docker/login .docker/build .docker/push
