package models

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"

	"blog/system"
)

// I don't need soft delete,so I use customized BaseModel instead gorm.Model
type BaseModel struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SmmsFile struct {
	BaseModel
	FileName  string `json:"filename"`
	StoreName string `json:"storename"`
	Size      int    `json:"size"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Hash      string `json:"hash"`
	Delete    string `json:"delete"`
	Url       string `json:"url"`
	Path      string `json:"path"`
}

func (sf SmmsFile) Insert() (err error) {
	err = DB.Create(&sf).Error
	return
}

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {

	db, err := gorm.Open("sqlite3", system.GetConfiguration().DSN)
	//db, err := gorm.Open("mysql", "root:mysql@/wblog?charset=utf8&parseTime=True&loc=Asia/Shanghai")
	if err == nil {
		DB = db
		//db.LogMode(true)
		db.AutoMigrate(&Page{}, &Post{}, &Tag{}, &PostTag{}, &User{}, &Comment{}, &Subscriber{}, &Link{}, &SmmsFile{})
		db.Model(&PostTag{}).AddUniqueIndex("uk_post_tag", "post_id", "tag_id")
		return db, err
	}
	return nil, err
}