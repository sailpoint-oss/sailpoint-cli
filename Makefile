.PHONY: clean
clean:
	go clean ./...

.PHONY: mocks
mocks:
	# Ref: https://github.com/golang/mock
	mockgen -source=internal/client/client.go -destination=internal/mocks/client.go -package=mocks
	mockgen -source=internal/terminal/terminal.go -destination=internal/mocks/terminal.go -package=mocks

.PHONY: test
test:
	go test -v -count=1 ./...

.PHONY: install
install:
	go build -o /usr/local/bin/sail -buildvcs=false

.PHONY: vhs
vhs:
	find assets -name "*.tape" | xargs -n 1 vhs

.PHONY: .docker/login .docker/build .docker/push
