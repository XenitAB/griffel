fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

#test: fmt vet lint
test: fmt vet
	go test --cover ./...

build:
	go build -o bin/griffel
