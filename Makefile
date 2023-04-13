.PHONY: clean
clean:
	go clean ./...

.PHONY: mocks
mocks:
	# Ref: https://github.com/golang/mock
	mockgen -source=client/client.go -destination=mocks/client.go -package=mocks

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
