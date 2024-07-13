build:
	go build \
		-ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(shell date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(shell git rev-parse HEAD)'" \
		-o cmd/agent/agent cmd/agent/main.go
	go build \
		-ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(shell date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(shell git rev-parse HEAD)'" \
		-o cmd/server/server cmd/server/main.go
	go build -o cmd/staticlint/staticlint cmd/staticlint/main.go
	go build -o cmd/keygen/keygen cmd/keygen/main.go
	# ./cmd/keygen/keygen
test:
	go test -v ./... -short
cover:
	rm -fr coverage
	mkdir coverage
	# флаг -count=1 не использовать кэш
	go test -count=1 ./... -coverprofile coverage/cover.out
	go tool cover -html coverage/cover.out -o coverage/cover.html
race:
	go test -v -race ./...
doc:
	godoc -http=:8088
vet:
	go vet -vettool=./cmd/staticlint/staticlint ./...
proto: 
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/proto/metric.proto

