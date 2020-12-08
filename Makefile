gen:
	@rm pb/*.go
	@protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:pb

clean:
	rm pb/*.go

run:
	go run main.go

test:
	@go test -cover -race ./...

server:
	@go run cmd/server/main.go -port 8080

client:
	@go run cmd/client/main.go -address 0.0.0.0:8080