package models

import (
	"github.com/cihub/seelog"
	. "blog/helpers"
	)

// table tags
type Tag struct {
	BaseModel
	Name  string         // tag name
	Total int `gorm:"-"` // count of post
}

// Tag
func (tag *Tag) Insert() error {
	return DB.FirstOrCreate(tag, "name = ?", tag.Name).Error
}

func (tag *Tag) Update() error {
	return DB.Save(tag).Error
}

func (tag *Tag) Delete() error {
	return DB.Delete(tag).Error
}

func ListTag() ([]*Tag, error) {
	var tags []*Tag
	rows, err := DB.Raw("select t.*,count(*) total from tags t inner join post_tags pt on t.id = pt.tag_id inner join posts p on pt.post_id = p.id where p.is_published = ? group by pt.tag_id", true).Rows()
	if err != nil {
		seelog.Error("[ListTag]db raw err", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag Tag
		DB.ScanRows(rows, &tag)
		tags = append(tags, &tag)
	}
	return tags, nil
}

func MustListTag() []*Tag {
	tags, _ := ListTag()
	return tags
}

func ListTagByPostId(id string) ([]*Tag, error) {
	var tags []*Tag
	pid, err := ParseIdToUint(id, "ListTagByPostId")
	if err != nil {
		return nil, err
	}
	rows, err := DB.Raw("select t.* from tags t inner join post_tags pt on t.id = pt.tag_id where pt.post_id = ?", uint(pid)).Rows()
	if err != nil {
		seelog.Error("[ListTagByPostId]db raw err", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag Tag
		DB.ScanRows(rows, &tag)
		tags = append(tags, &tag)
	}
	return tags, nil
}

func CountTag() int {
	var count int
	DB.Model(&Tag{}).Count(&count)
	return count
}

func ListAllTag() ([]*Tag, error) {
	var tags []*Tag
	err := DB.Model(&Tag{}).Find(&tags).Error
	return tags, err
}

