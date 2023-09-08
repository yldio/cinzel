COVERAGE_PATH=./coverage
COVERAGE_REPORT_PATH=$(COVERAGE_PATH)/coverage.out
COVERAGE_HTML_PATH=$(COVERAGE_PATH)/coverage.html
BINARY=atos

run:
ifdef file
	@go run ./cmd/main.go --file=$(file)
endif
ifdef dir
	@go run ./cmd/main.go --dir=$(dir)
endif

build:
	@go build -ldflags "-s -w" -o ./bin/$(BINARY) ./cmd/main.go
	@wc -c ./bin/$(BINARY)

fmt:
	@go fmt ./...
	@terragrunt hclfmt

test-ci:
	@go test ./... --cover

test:
	@go test ./... --cover -coverprofile=$(COVERAGE_REPORT_PATH)

test-cover:
	@go tool cover -html=$(COVERAGE_REPORT_PATH) -o $(COVERAGE_HTML_PATH)

update-report: test test-cover

cover-ui: update-report
	@open $(COVERAGE_HTML_PATH)

changelog-create:
	@git cliff --output CHANGELOG.md