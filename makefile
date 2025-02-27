ifneq (,$(wildcard ./.env))
    include .env
    export
endif

COVERAGE_PATH=./coverage
COVERAGE_REPORT_PATH=$(COVERAGE_PATH)/coverage.out
COVERAGE_HTML_PATH=$(COVERAGE_PATH)/coverage.html

check_version:
ifndef VERSION
	$(error VERSION is undefined)
endif

fmt:
	@go fmt ./...
	@go tool hclfmt -w ./**/*.hcl

lint:
	@go tool golangci-lint run

build: check_version
	@go build -ldflags "-s -w -X 'main.version=${VERSION}'" -o ./bin/$(BINARY) ./cmd/$(BINARY)/main.go

test-ci:
	@go test ./... --cover

test:
	@go test ./... --cover -coverprofile=$(COVERAGE_REPORT_PATH)

test-cover:
	@go tool cover -html=$(COVERAGE_REPORT_PATH) -o $(COVERAGE_HTML_PATH)

update-report: test test-cover

cover-ui: update-report
	@open $(COVERAGE_HTML_PATH)

docs-ui:
	@open http://localhost:6060/
	@godoc -http=:6060

changelog:
	go tool changie

actions:
	@./bin/$(BINARY) --directory=./$(BINARY) --output-directory=.github/workflows

docker-build: check_version
	@docker build --build-arg version=${VERSION} -t $(BINARY) .

docker-run:
	@docker run --rm -it $(BINARY) --version