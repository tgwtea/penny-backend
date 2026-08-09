.PHONY: build check-contract check-generated db-check fmt generate lint test test-integration verify vuln

build:
	go build ./cmd/api ./cmd/worker ./cmd/migrate

generate:
	go generate ./...

fmt:
	go tool golangci-lint fmt

lint:
	go vet ./...
	go tool golangci-lint run

test:
	go test -race -coverprofile=coverage.out ./...

test-integration:
	go test -race -tags=integration ./...

db-check:
	go run ./cmd/dbcheck

check-generated: generate
	git diff --exit-code -- openapi.yaml internal/api/openapi.generated.go internal/platform/database/dbgen

check-contract:
	bash .github/scripts/check-contract.sh

vuln:
	go tool govulncheck ./...

verify: check-generated lint test vuln build

