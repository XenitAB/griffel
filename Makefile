fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test: fmt vet lint
	go test --cover ./...

build:
	go build -o bin/griffel
