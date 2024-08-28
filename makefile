COVERAGE_PATH=./coverage
COVERAGE_REPORT_PATH=$(COVERAGE_PATH)/coverage.out
COVERAGE_HTML_PATH=$(COVERAGE_PATH)/coverage.html
BINARY=acto

run:
ifdef file
	@go run ./cmd/$(BINARY)/main.go --file=$(file)
endif
ifdef dir
	@go run ./cmd/$(BINARY)/main.go --dir=$(dir)
endif

build:
	@rm -rf ./dist
	@goreleaser release --snapshot --skip=publish

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

docs-ui:
	@open http://localhost:6060/
	@godoc -http=:6060

changelog:
	@git cliff --output CHANGELOG.md