.PHONY: build test vet fmt lint snapshot clean

build:
	go build -o bin/secretcheck ./cmd/secretcheck

test:
	go test ./... -v

vet:
	go vet ./...

fmt:
	gofmt -l .

# Cross-compiles for all platforms into dist/ without publishing anything.
snapshot:
	goreleaser build --snapshot --clean

clean:
	rm -rf bin dist
