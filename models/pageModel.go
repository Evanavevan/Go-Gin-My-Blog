package models

import . "blog/helpers"

// table pages
type Page struct {
	BaseModel
	Title       string // title
	Body        string // body
	View        int    // view count
	IsPublished bool   // published or not
}

func (page *Page) Insert() error {
	return DB.Create(page).Error
}

func (page *Page) Update() error {
	return DB.Model(page).Updates(map[string]interface{}{
		"title":        page.Title,
		"body":         page.Body,
		"is_published": page.IsPublished,
	}).Error
}

func (page *Page) UpdateView() error {
	return DB.Model(page).Updates(map[string]interface{}{
		"view": page.View,
	}).Error
}

func (page *Page) Delete() error {
	return DB.Delete(page).Error
}

func GetPageById(id string) (*Page, error) {
	pid, err := ParseIdToUint(id, "GetPageById")
	if err != nil {
		return nil, err
	}
	var page Page
	err = DB.First(&page, "id = ?", pid).Error
	return &page, err
}

func ListPublishedPage() ([]*Page, error) {
	return _listPage(true)
}

func ListAllPage() ([]*Page, error) {
	return _listPage(false)
}

func _listPage(published bool) ([]*Page, error) {
	var pages []*Page
	var err error
	if published {
		err = DB.Where("is_published = ?", true).Find(&pages).Error
	} else {
		err = DB.Find(&pages).Error
	}
	return pages, err
}

func CountPage() int {
	var count int
	DB.Model(&Page{}).Count(&count)
	return count
}