clean:
	go clean ./...

mocks:
	# Ref: https://github.com/golang/mock
	mockgen -source=client/client.go -destination=mocks/client.go -package=mocks

test:
	go test -v -count=1 ./...

install:
	go build -o /usr/local/bin/sail -buildvcs=false

vhs-auto:
	vhs < assets/vhs/brewinstall.tape
	vhs < assets/vhs/linuxMake.tape
	vhs < assets/vhs/sail.tape
	vhs < assets/vhs/configure/configure-pat.tape
	vhs < assets/vhs/configure/configure-oauth.tape
	vhs < assets/vhs/va/va-collect.tape
	vhs < assets/vhs/va/va-update.tape
	vhs < assets/vhs/va/va-parse.tape
	vhs < assets/vhs/va/va-troubleshoot.tape
	vhs < assets/vhs/transform/transform-list.tape
	vhs < assets/vhs/transform/transform-download.tape


.PHONY: clean mocks test install vhs-auto .docker/login .docker/build .docker/push
