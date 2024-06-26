package database

import (
	"fmt"
	"log"

	"github.com/namnv2496/book_service/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	// "gorm.io/gorm/logger"
)

type UserRepository struct {
	gorm.Model
}

var (
	db *gorm.DB
)

func DBConnect() {

	config := config.ReadConfig(config.DB).(config.DatabaseConfig)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
	)
	// dsn := config.DB.Username + ":" + config.DB.Password + "@tcp" + "(" + config.DB.Host + ":" + strconv.Itoa(config.DB.Port) + ")/" + config.DB.Name + "?" + "parseTime=true&loc=Local"

	log.Println(dsn)
	d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		fmt.Print("connect to DB fail")
		panic(err)
	}
	fmt.Println("Connect to DB done")
	db = d
}

func GetDB() *gorm.DB {
	return db
}
