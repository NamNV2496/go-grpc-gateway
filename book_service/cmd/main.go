package main

import (
	"log"
	"net"

	"github.com/namnv2496/book_service/domain"
	"github.com/namnv2496/book_service/generated/pb"
	"github.com/namnv2496/book_service/internal/config"
	"github.com/namnv2496/book_service/internal/database"
	server "github.com/namnv2496/book_service/internal/hanlder/grpc"
	"github.com/namnv2496/book_service/internal/logic/auth"
	"google.golang.org/grpc"
)

func main() {
	config.InitReadConfig()
	database.DBConnect()

	if err := database.GetDB().AutoMigrate(&domain.Book{}); err != nil {
		log.Fatalf("Failed to migrate book table in database")
	}
	if err := database.GetDB().AutoMigrate(&domain.Author{}); err != nil {
		log.Fatalf("Failed to migrate author table in database")
	}
	if err := runGrpcServer(); err != nil {
		log.Fatalf("Failed to start grpc server")
	}
}

func runGrpcServer() error {

	serverGrpc := config.ReadConfig(config.Book).(config.ServerConfig)

	interceptor := auth.NewAuthInterceptor()
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Unary()),
	)
	listener, err := net.Listen("tcp", serverGrpc.Address)
	if err != nil {
		return err
	}
	handler := server.NewBookServer(database.GetDB())
	pb.RegisterBookServiceServer(grpcServer, handler)
	log.Println("run grpc server in port ", serverGrpc.Address)
	return grpcServer.Serve(listener)
}
