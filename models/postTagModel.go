package models

// table post_tags
type PostTag struct {
	BaseModel
	PostId uint // post id
	TagId  uint // tag id
}

// post_tags
func (pt *PostTag) Insert() error {
	return DB.FirstOrCreate(pt, "post_id = ? and tag_id = ?", pt.PostId, pt.TagId).Error
}

func DeletePostTagByPostId(postId uint) error {
	return DB.Delete(&PostTag{}, "post_id = ?", postId).Error
}
