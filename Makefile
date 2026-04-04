.PHONY: run tidy build test

# Start everything in one terminal
run:
	go run cmd/server/main.go

tidy:
	go mod tidy

build:
	go build -o bin/server.exe cmd/server/main.go

test:
	curl -X POST http://localhost:8080/payments \
		-H "Content-Type: application/json" \
		-d "{\"user_id\": \"user_123\", \"amount\": 9000, \"idempotency_key\": \"key_$(shell date +%s)\"}"