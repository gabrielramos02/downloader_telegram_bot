.PHONY: all tools update-deps fmt security staticcheck test check pre-commit clean

all: check

tools:
	@which gosec > /dev/null || (echo "gosec not installed"; exit 1)
	@which staticcheck > /dev/null || (echo "staticcheck not installed"; exit 1)

update-deps:
	go get github.com/gabrielramos02/gopeed-api-go@main
	go mod tidy

fmt:
	go fmt ./...

security:
	gosec ./...

staticcheck:
	staticcheck ./...

test:
	go test ./...

check: tools fmt security staticcheck test

pre-commit: update-deps check

clean:
	rm -rf tmp/ telegram-bot
