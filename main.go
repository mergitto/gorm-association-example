package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func main() {
	db, err := newDB()
	if err != nil {
		fmt.Printf("%+v", err)
		return
	}
	if err := migrate(db); err != nil {
		fmt.Printf("%+v", err)
		return
	}
	if err := seeds(db); err != nil {
		fmt.Printf("%+v", err)
		return
	}
	fmt.Println("【Book, has one Publisher, many2many Author】")
	if err := getBook(db); err != nil {
		fmt.Printf("%+v", err)
		return
	}
	fmt.Println("【Author, many2many Book】")
	if err := getAuthor(db); err != nil {
		fmt.Printf("%+v", err)
		return
	}
	fmt.Println("【Publisher, has many Book】")
	if err := getPublisher(db); err != nil {
		fmt.Printf("%+v", err)
		return
	}
}

func getBook(db *gorm.DB) error {
	var books []Book
	if err := db.Preload("Publisher").Preload("Authors").Find(&books).Error; err != nil {
		return err
	}
	for _, v := range books {
		fmt.Printf("=========== %s ==========\n", v.Title)
		fmt.Printf("book: {id: %d, title: %s}\n", v.ID, v.Title)
		fmt.Printf("publisher: {id: %d, name: %s}\n", v.Publisher.ID, v.Publisher.Name)
		for i, a := range v.Authors {
			fmt.Printf("authors-%d: {id: %d, name: %s}\n", i, a.ID, a.Name)
		}
	}
	fmt.Println()
	return nil
}

func getAuthor(db *gorm.DB) error {
	var authors []Author
	if err := db.Preload("Books").Find(&authors).Error; err != nil {
		return nil
	}
	for _, v := range authors {
		fmt.Printf("=========== %s ==========\n", v.Name)
		fmt.Printf("author: {id: %d, name: %s}\n", v.ID, v.Name)
		for i, b := range v.Books {
			fmt.Printf("books-%d: {id: %d, title: %s}\n", i, b.ID, b.Title)
		}
	}
	fmt.Println()
	return nil
}

func getPublisher(db *gorm.DB) error {
	var publishers []Publisher
	if err := db.Preload("Books").Find(&publishers).Error; err != nil {
		return err
	}
	for _, v := range publishers {
		fmt.Printf("=========== %s ==========\n", v.Name)
		fmt.Printf("pulisher: {id: %d, name: %s}\n", v.ID, v.Name)
		for i, b := range v.Books {
			fmt.Printf("books-%d: {id: %d, title: %s}\n", i, b.ID, b.Title)
		}
	}
	fmt.Println()
	return nil
}

type Book struct {
	ID          uint   `gorm:"primary_key;AUTO_INCREMENT;not null;"`
	Title       string `gorm:"not null"`
	PublisherID uint   `gorm:"not null"`
	Publisher   Publisher
	Authors     []Author `gorm:"many2many:author_books"`
}

type Author struct {
	ID    uint   `gorm:"primary_key;AUTO_INCREMENT;not null;"`
	Name  string `gorm:"not null"`
	Books []Book `gorm:"many2many:author_books"`
}

type Publisher struct {
	ID    uint   `gorm:"primary_key;AUTO_INCREMENT;not null;"`
	Name  string `gorm:"not null"`
	Books []Book
}

func newDB() (*gorm.DB, error) {
	dbconf := fmt.Sprintf("%s:%s@%s(%s)/%s",
		"root",
		"mysql",
		"",
		"localhost",
		"play-ground",
	) + "?parseTime=true&loc=Asia%2FTokyo&charset=utf8mb4"
	db, err := gorm.Open("mysql", dbconf)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func migrate(db *gorm.DB) error {
	// 外部キーも設定すべきですが無視します
	if err := db.AutoMigrate(&Book{}).
		AutoMigrate(&Author{}).
		AutoMigrate(&Publisher{}).
		Error; err != nil {
		return err
	}
	return nil
}

const (
	publisherName = "test-publisher"
	authorName1   = "test-author-1"
	authorName2   = "test-author-2"
	BookTitle1    = "test-book-1"
	BookTitle2    = "test-book-2"
)

func seeds(db *gorm.DB) error {
	if !db.First(&Publisher{Name: publisherName}).RecordNotFound() {
		return nil
	}

	publisher := Publisher{Name: publisherName}
	if err := db.Create(&publisher).Error; err != nil {
		return err
	}

	author1 := Author{Name: authorName1}
	author2 := Author{Name: authorName2}
	if err := db.Create(&author1).Create(&author2).Error; err != nil {
		return err
	}

	// book を作成 & 中間テーブルに作成
	book1 := Book{Title: BookTitle1, PublisherID: publisher.ID}
	book2 := Book{Title: BookTitle2, PublisherID: publisher.ID}
	if err := db.Model(&author1).Association("Books").Append(&book1).Append(&book2).Error; err != nil {
		return err
	}
	if err := db.Model(&author2).Association("Books").Append(&book2).Error; err != nil {
		return err
	}
	return nil
}
