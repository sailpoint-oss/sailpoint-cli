SCOPES_SNAPSHOT := internal/scopes/data/sailpoint-scopes.json

.PHONY: clean
clean:
	go clean ./...

.PHONY: mocks
mocks:
	# Ref: https://github.com/golang/mock
	mockgen -source=internal/client/client.go -destination=internal/mocks/client.go -package=mocks
	mockgen -source=internal/terminal/terminal.go -destination=internal/mocks/terminal.go -package=mocks

.PHONY: test
test: $(SCOPES_SNAPSHOT)
	go test -v -count=1 ./...

.PHONY: test-report
test-report:
	@echo "Running tests..."
	@go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest
	@mkdir -p test-results
	@bash -c "go test -json -v -coverprofile=coverage.txt -covermode=atomic ./... 2>&1 | tee test-results/gotest.log | gotestfmt"
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.txt -o test-results/coverage.html
	@echo "Test results and coverage saved in test-results directory"

.PHONY: test-race
test-race:
	@echo "Running tests with race detection enabled..."
	@CGO_ENABLED=1 go test -v -race ./...

# Generates the scope snapshot only when it is missing.
$(SCOPES_SNAPSHOT):
	go generate ./internal/scopes/...

# Always downloads the latest spec and regenerates the scope snapshot.
.PHONY: generate
generate:
	go generate ./internal/scopes/...

.PHONY: install
install: $(SCOPES_SNAPSHOT)
	go build -o /usr/local/bin/sail -buildvcs=false

.PHONY: vhs
vhs:
	find assets -name "*.tape" | xargs -n 1 vhs

.PHONY: .docker/login .docker/build .docker/push
