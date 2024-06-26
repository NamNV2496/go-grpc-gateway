package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/namnv2496/fe_service/generated/pb"
	"github.com/namnv2496/fe_service/internal/config"
	"github.com/namnv2496/fe_service/internal/logic/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {

	config.InitReadConfig()

	// Example for FE or another service
	if err := runGrpcClient(); err != nil {
		log.Fatalf("Failed to start server grpc")
	}
}

const (
	username        = "user1"
	password        = "secret"
	refreshDuration = 24 * time.Hour
)

func authMethods() map[string]bool {
	const bookServicePath = "/book.pb.BookService/"

	return map[string]bool{
		bookServicePath + "CreateNewBook": true,
	}
}

func runGrpcClient() error {
	serverAddress := flag.String("address", "", "the server address")
	enableTLS := flag.Bool("tls", false, "enable SSL/TLS")
	flag.Parse()

	authEndpoint := config.ReadConfig(config.Server).(config.ServerConfig).Address
	transportOption := grpc.WithInsecure()
	if *enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			log.Fatal("cannot load TLS credentials: ", err)
		}
		transportOption = grpc.WithTransportCredentials(tlsCredentials)
	}
	var address string
	if *serverAddress == "" {
		address = authEndpoint
	} else {
		address = *serverAddress
	}
	log.Println("address: ", address)
	// demo of login
	// Creat first connection to get token from http_gateway service
	cc1, err := grpc.Dial(address, transportOption)
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	authClient := client.NewAuthClient(cc1, username, password)
	interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration)
	if err != nil {
		log.Fatal("cannot create auth interceptor: ", err)
	}

	// main connection for requesting to http_gateway service
	_, err = grpc.Dial(
		address,
		transportOption,
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	// connect to book_server
	bookEndpoint := config.ReadConfig(config.Book).(config.ServerConfig).Address
	bookConn, err := grpc.Dial(
		bookEndpoint,
		transportOption,
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	grpcClient := pb.NewBookServiceClient(bookConn)

	// testcases
	testCreateNewBook(grpcClient)
	testUploadImage(grpcClient)
	return nil
}

func testCreateNewBook(grpcClient pb.BookServiceClient) {
	request := pb.BookRequest{
		Name:       "hoa cỏ may",
		Price:      100000,
		PublicDate: "2024-06-25 18:00:00",
		AuthorId:   1,
	}
	log.Println("Call create new book")
	bookId, _ := grpcClient.CreateNewBook(context.Background(), &request)

	log.Println("bookId: ", bookId)
}

func testUploadImage(grpcClient pb.BookServiceClient) {

	imagePath := "file/1.png"
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()
	// ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	// defer cancel()
	ctx := context.Background()

	stream, err := grpcClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				ImageType: filepath.Ext(imagePath),
				Name:      "1.png",
			},
		},
	}
	// send info first
	log.Println("sent information first")
	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info to server: ", err, stream.RecvMsg(nil))
	}
	// send data of image
	log.Println("sent data")
	reader := bufio.NewReader(file)
	buffer := make([]byte, 100000)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image uploaded with message: %s, size: %d", res.GetMessage(), res.GetSize())
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := os.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}
