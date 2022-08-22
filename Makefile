
test:
	go test -coverprofile=test-coverage.out ./...

coverage:
	go tool cover -html="test-coverage.out" -o="test-coverage.html"

build:
	go build -o keyring .