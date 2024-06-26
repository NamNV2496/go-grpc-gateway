package domain

import (
	"github.com/namnv2496/book_service/generated/pb"
	"gorm.io/gorm"
)

type Author struct {
	gorm.Model
	Name       string
	Age        int32
	Experience int32
	Opinion    string
	Sex        int
}

func (u Author) TableName() string {
	return "author"
}

func GetAuthorByName(name string, db *gorm.DB) (*Author, error) {

	var author Author
	err := db.Where("name = ?", name).First(&author).Error
	if err != nil {
		return nil, err
	}
	return &author, nil
}

func CreateAuthor(req *pb.AuthorRequest, db *gorm.DB) (int32, error) {

	author := Author{
		Name:       req.Name,
		Age:        req.Age,
		Experience: req.Experience,
		Opinion:    req.Opinion,
	}

	result := db.Create(&author)
	return int32(author.ID), result.Error
}
