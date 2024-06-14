build:
	go build -o cmd/agent/agent cmd/agent/main.go
	go build -o cmd/server/server cmd/server/main.go
	go build -o cmd/staticlint/staticlint cmd/staticlint/main.go
test:
	go test -v ./... -short
cover:
	rm -fr coverage
	mkdir coverage
	# флаг -count=1 не использовать кэш
	# флаг -p 1 запуск тестов в одном потоке (в Example Агента и Сервера для запуска тестов используется один и тот же endpoint[:8080 и :8081] соответственно)
	go test -count=1 -p 1 ./... -coverprofile coverage/cover.out
	go tool cover -html coverage/cover.out -o coverage/cover.html
race:
	go test -v -race ./...
doc:
	godoc -http=:8088
vet:
	go vet -vettool=./cmd/staticlint/staticlint ./...
