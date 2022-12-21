clean:
	go clean ./...

mocks:
	# Ref: https://github.com/golang/mock
	mockgen -source=client/client.go -destination=mocks/client.go -package=mocks

test:
	go test -v -count=1 ./...

install:
	go build -o /usr/local/bin/sail -buildvcs=false

vhs:
	vhs < assets/vhs/linuxMake.tape
	vhs < assets/vhs/sail.tape
	vhs < assets/vhs/configure-pat.tape
	vhs < assets/vhs/configure-oauth.tape


.PHONY: clean mocks test install vhs .docker/login .docker/build .docker/push
