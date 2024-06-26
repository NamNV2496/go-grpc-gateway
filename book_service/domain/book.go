package domain

import (
	"github.com/namnv2496/book_service/generated/pb"
	"gorm.io/gorm"
)

type Book struct {
	gorm.Model
	Name       string
	Price      int32
	PublicDate string
	AuthorId   int32
}

func (u Book) TableName() string {
	return "book"
}

func GetBookByName(name string, db *gorm.DB) (*Book, error) {

	var book Book
	err := db.Where("name=?", name).Find(&book).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func CreateBook(req *pb.BookRequest, db *gorm.DB) (int32, error) {

	book := Book{
		Name:       req.Name,
		Price:      req.Price,
		PublicDate: req.PublicDate,
		AuthorId:   req.AuthorId,
	}

	result := db.Create(&book)
	return int32(book.ID), result.Error
}
