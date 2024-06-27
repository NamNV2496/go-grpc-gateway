package server

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/namnv2496/book_service/domain"
	"github.com/namnv2496/book_service/generated/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type BookServer struct {
	pb.UnimplementedBookServiceServer
	db *gorm.DB
}

func NewBookServer(
	db *gorm.DB,
) pb.BookServiceServer {
	return &BookServer{
		db: db,
	}
}

func (bs *BookServer) CreateNewAuthor(
	ctx context.Context,
	req *pb.AuthorRequest,
) (*pb.AuthorResponse, error) {

	log.Println("call create new author")
	author, err := domain.GetAuthorByName(req.Name, bs.db)
	if err != nil {
		return nil, err
	}
	if author != nil {
		log.Println("Author %w was existed", &author.Name)
		return &pb.AuthorResponse{
			Id: int32(author.ID),
		}, nil
	}
	log.Println("Add new author")
	authorId, err := domain.CreateAuthor(req, bs.db)
	if err != nil {
		return nil, err
	}
	return &pb.AuthorResponse{
		Id: authorId,
	}, nil
}

func (bs *BookServer) CreateNewBook(
	ctx context.Context,
	req *pb.BookRequest,
) (*pb.BookResponse, error) {
	log.Println("call create new book")
	book, err := domain.GetBookByName(req.Name, bs.db)
	if err != nil {
		return nil, err
	}
	if book != nil {
		log.Println("Book was existed: ", &book.Name)
		return &pb.BookResponse{
			Id: int32(book.ID),
		}, nil
	}
	bookId, err := domain.CreateBook(req, bs.db)
	if err != nil {
		return nil, err
	}
	return &pb.BookResponse{
		Id: bookId,
	}, nil
}

const maxImageSize = 1 << 20

func (bs *BookServer) UploadImage(stream pb.BookService_UploadImageServer) error {
	// receive information of image first
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive image info"))
	}
	imageType := req.GetInfo().GetImageType()
	imageName := req.GetInfo().GetName()
	log.Println("Received request of image: ", imageName, " type: ", imageType)

	imageSize := 0
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}
		log.Print("waiting to receive more data")
		// receive data of image
		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			// response receive done
			stream.SendAndClose(&pb.UploadImageResponse{
				Message: "done",
				Size:    uint32(imageSize),
			})
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunkData()
		size := len(chunk)
		log.Printf("received a chunk with size: %d", size)

		imageSize += size
		if imageSize > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize))
		}

		// write slowly
		time.Sleep(time.Second * 5)

		// _, err = imageData.Write(chunk)
		err = os.WriteFile("./internal/handler/tmp/"+imageName, chunk, 0664)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}

func (bs *BookServer) SearchBook(
	req *pb.BookRequest,
	res pb.BookService_SearchBookServer,
) error {
	return nil
}
