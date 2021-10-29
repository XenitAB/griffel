fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test: fmt vet
	go test --cover ./...

build:
	go build -o bin/griffel

e2e: build
	./test/e2e/e2e.sh
