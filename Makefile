VERSION ?= dev

clean:
	go clean ./...

mocks:
	# Ref: https://github.com/golang/mock
	mockgen -source=client/client.go -destination=mocks/client.go -package=mocks

test:
	docker build -t cli .
	docker run --rm cli go test -v -count=1 ./...

install:
	go build -o /usr/local/bin/sp

docker/login:
ifeq ($(JENKINS_URL),) # if $JENKINS_URL is empty
	aws ecr --region us-east-1 get-login-password | docker login --username AWS --password-stdin 406205545357.dkr.ecr.us-east-1.amazonaws.com
else
	$$(aws ecr get-login --no-include-email --region us-east-1)
endif

docker/build: docker/login
	docker build -t sailpoint/sp-cli:$(VERSION) -f Dockerfile .

docker/push: docker/build
	docker tag sailpoint/sp-cli:$(VERSION) 406205545357.dkr.ecr.us-east-1.amazonaws.com/sailpoint/sp-cli:$(VERSION)
	docker push 406205545357.dkr.ecr.us-east-1.amazonaws.com/sailpoint/sp-cli:$(VERSION)

.PHONY: clean mocks test install .docker/login .docker/build .docker/push
