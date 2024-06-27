package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/namnv2496/http_gateway/generated/pb"
	"github.com/namnv2496/http_gateway/internal/config"
	"github.com/namnv2496/http_gateway/internal/database"
	"github.com/namnv2496/http_gateway/internal/domain"
	"github.com/namnv2496/http_gateway/internal/logic/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	secretKey     = "secretKey"
	tokenDuration = time.Hour * 1
)

func InitDatabase() {
	database.DBConnect()
	if err := database.GetDB().AutoMigrate(&domain.User{}); err != nil {
		log.Fatalln("Failed to migrate database")
	}
}

func accessibleRoles() map[string][]string {
	// const authServicePath = "/user.pb.AuthService/"

	return map[string][]string{
		// authServicePath + "test": {"admin"},
		// authServicePath + "UploadImage":  {"admin"},
		// authServicePath + "RateLaptop":   {"admin", "user"},
	}
}

func main() {
	serverType := flag.String("type", "grpc", "type of server (grpc/rest)")
	serverPort := flag.Int("port", 5600, "port of server")
	enableTLS := flag.Bool("tls", false, "enable SSL/TLS")
	flag.Parse()

	config.InitReadConfig()
	if *serverType == "grpc" {
		InitDatabase()
		jwtManner := auth.NewJWTManager(secretKey, tokenDuration)
		authService := auth.NewAuthServer(jwtManner, database.GetDB())
		if err := runGrpcServer(authService, jwtManner, *enableTLS, *serverPort); err != nil {
			log.Fatalf("Failed to start grpc server: %s", err)
		}
	} else {
		if err := runRESTServer(*serverPort); err != nil {
			log.Fatalf("Failed to start http server: %s", err)
		}
	}
}

func runGrpcServer(
	authServer pb.AuthServiceServer,
	jwtManager *auth.JWTManager,
	enableTLS bool,
	serverPort int,
) error {
	interceptor := auth.NewAuthInterceptor(jwtManager, accessibleRoles())

	// add interceptor for server
	var grpcServer *grpc.Server
	if enableTLS {
		tlsCredential, err := loadTLSCredential()
		if err != nil {
			return err
		}
		grpcServer = grpc.NewServer(
			grpc.Creds(tlsCredential),
			grpc.UnaryInterceptor(interceptor.Unary()),
			grpc.StreamInterceptor(interceptor.Stream()),
		)
	} else {
		grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(interceptor.Unary()),
			grpc.StreamInterceptor(interceptor.Stream()),
		)
	}
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	address := fmt.Sprintf("0.0.0.0:%d", serverPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	log.Println("run grpc server in port ", address, ", tls = ", enableTLS)
	return grpcServer.Serve(listener)
}

func runRESTServer(serverPort int) error {

	httpEndpoint := config.ReadConfig(config.Server).(config.ServerConfig).Address
	bookEndpoint := config.ReadConfig(config.Book).(config.ServerConfig).Address
	mux := runtime.NewServeMux()
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, httpEndpoint, dialOptions)
	if err != nil {
		return err
	}
	err = pb.RegisterBookServiceHandlerFromEndpoint(ctx, mux, bookEndpoint, dialOptions)
	if err != nil {
		return err
	}

	log.Println("Start REST server at port: ", serverPort)
	address := fmt.Sprintf("0.0.0.0:%d", serverPort)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("Failed to serve: %v", err)
		return err
	}
	return nil
}

const (
	serverCertFile   = "cert/server-cert.pem"
	serverKeyFile    = "cert/server-key.pem"
	clientCACertFile = "cert/ca-cert.pem"
)

func loadTLSCredential() (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, err
	}

	pemClientCA, err := os.ReadFile(clientCACertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}
