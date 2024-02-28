build: go-imports
	@rm -rf bin/
	go build -o bin/ ./...

gh-build:
	@rm -rf bin/
	go build -o bin/ ./...

go-imports:
	goimports -w .

upgrade-deps:
	go get -u ./...
	go mod tidy
	gotestsum ./...

lint:
	golangci-lint run

gh-test:
	go test ./...

test:
	gotestsum ./...
