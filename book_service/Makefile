gen:
	protoc --proto_path=../api/proto ../api/proto/*.proto  --go_out=:generated/pb --go-grpc_out=:generated/pb --grpc-gateway_out=:generated/pb --openapiv2_out=:generated/swagger


server:
	go run cmd/main.go

lint:
	golangci-lint run