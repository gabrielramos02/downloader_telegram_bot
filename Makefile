.PHONY: all fmt-check update-deps fmt style tests security test-cover staticcheck test check pre-commit clean 

all: check

# Style
fmt-check:
	@test -z "$$(go fmt ./...)"

staticcheck:
	@which staticcheck > /dev/null || (echo "staticcheck not installed"; exit 1)
	staticcheck ./...

style: fmt-check staticcheck
# Test 
security:
	@which gosec > /dev/null || (echo "gosec not installed"; exit 1)
	gosec ./...

test-cover:
	go test ./... -cover

tests : security test-cover

fmt:
	go fmt ./...
test:
	go test ./...

check: staticcheck security fmt test 

update-deps:
	go get github.com/gabrielramos02/gopeed-api-go@main
	go mod tidy

pre-commit: update-deps check

clean:
	rm -rf tmp/ telegram-bot
