gen:
	protoc --proto_path=../api/proto ../api/proto/*.proto  --go_out=:generated/pb --go-grpc_out=:generated/pb --grpc-gateway_out=:generated/pb --openapiv2_out=:generated/swagger

rest:
	go run cmd/main.go -type rest

rest-tls:
	go run cmd/main.go -type rest -tls true

server:
	go run cmd/main.go

server-tls:
	go run cmd/main.go -tls

server-tls-nginx1:
	go run cmd/main.go -port 50051 -tls

server-tls-nginx2:
	go run cmd/main.go -port 50052 -tls

cert:
	cd cert; ./gen.sh; cd ..
lint:
	golangci-lint run
