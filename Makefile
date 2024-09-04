VERSION = v1.0.0
COMMIT = $(shell git rev-list -1 HEAD | cut -c1-10)
DIRTY = $(shell git diff --quiet || echo '-dirty')
DATE = $(shell date -u +%a,\ %d\ %b\ %Y\ %H:%M:%S\ %Z)

all: test build

clean:
	go mod tidy
	go fmt ./...
	rm -f netbox_sd || true
	rm -f coverage.out || true

test:
	go vet ./...
	# https://staticcheck.io/
	#staticcheck -go 1.23 ./...
	# https://go.dev/blog/vuln
	#govulncheck ./...
	go clean -testcache && go test -short -cover -coverprofile=coverage-short.out -covermode=set ./...
	go tool cover -func=coverage-short.out

integration:
	go clean -testcache && go test -cover -coverprofile=coverage.out -covermode=set ./...
	go tool cover -func=coverage.out

build:
	go build -race -ldflags '-X main.version=${VERSION} -X main.commit=${COMMIT}${DIRTY} -X "main.date=${DATE}"' -o bin/netbox_sd
