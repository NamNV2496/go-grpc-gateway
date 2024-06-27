gen:
	protoc --proto_path=../api/proto ../api/proto/*.proto  --go_out=:generated/pb --go-grpc_out=:generated/pb --grpc-gateway_out=:generated/pb --openapiv2_out=:generated/swagger

client:
	go run cmd/main.go

client-tls:
	go run cmd/main.go -tls

client-tls-nginx:
	go run cmd/main.go -address 0.0.0.0:8081 -tls

cert:
	cd cert/; ./gen.sh; cd ..

lint:
	golangci-lint run