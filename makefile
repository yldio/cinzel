ifneq (,$(wildcard ./.env))
    include .env
    export
endif

check_binary:
ifndef BINARY
	$(error BINARY is undefined)
endif

check_version:
ifndef VERSION
	$(error VERSION is undefined)
endif

do_checks: check_binary check_version

fmt:
	@go fmt ./...
	@go tool hclfmt -w ./**/*.hcl

lint:
	@go tool golangci-lint run

build: do_checks
	rm -rf ./bin
	go build -ldflags "-s -w -X main.version=${VERSION}" -o ./bin/$(BINARY) ./$(BINARY).go

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

docker-build: do_checks
	@docker build --file=./build/Dockerfile --build-arg version=${VERSION} -t "$(BINARY):${VERSION}" .

docker-run: do_checks
	@docker run --rm -it $(BINARY) --version