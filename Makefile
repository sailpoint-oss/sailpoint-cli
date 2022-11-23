clean:
	go clean ./...

mocks:
	# Ref: https://github.com/golang/mock
	mockgen -source=client/client.go -destination=mocks/client.go -package=mocks

test:
	go test -v -count=1 ./...

install:
	go build -o /usr/local/bin/sail -buildvcs=false

vhslinux:
	vhs < vhs/linuxMake.tape
	vhs < vhs/sail.tape

vhswindows:
	echo "not yet configured"

.PHONY: clean mocks test install vhslinux vhswindows .docker/login .docker/build .docker/push
