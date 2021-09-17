package models

import (
	"github.com/cihub/seelog"
	. "blog/helpers"
)

// table comments
type Comment struct {
	BaseModel
	UserID    uint                      // 用户id
	Content   string                    // 内容
	PostID    uint                      // 文章id
	ReadState bool `gorm:"default:'0'"` // 阅读状态
	//Replies []*Comment // 评论
	NickName      string `gorm:"-"`
	AvatarUrl     string `gorm:"-"`
	GithubUrl     string `gorm:"-"`
	githubLoginId string
}

// Comment
func (comment *Comment) Insert() error {
	return DB.Create(comment).Error
}

func (comment *Comment) Update() error {
	return DB.Model(comment).UpdateColumn("read_state", true).Error
}

func SetAllCommentRead() error {
	return DB.Model(&Comment{}).Where("read_state = ?", false).Update("read_state", true).Error
}

func ListUnreadComment() ([]*Comment, error) {
	var comments []*Comment
	err := DB.Where("read_state = ?", false).Order("created_at desc").Find(&comments).Error
	return comments, err
}

func MustListUnreadComment() []*Comment {
	comments, _ := ListUnreadComment()
	return comments
}

func (comment *Comment) Delete() error {
	return DB.Delete(comment, "user_id = ?", comment.UserID).Error
}

func ListCommentByPostID(postId string) ([]*Comment, error) {
	pid, err := ParseIdToUint(postId, "ListCommentByPostID")
	if err != nil {
		return nil, err
	}
	var comments []*Comment
	rows, err := DB.Raw("select c.*, u.github_login_id, u.nick_name, u.avatar_url, u.github_url from comments c inner join users u on c.user_id = u.id where c.post_id = ? order by created_at desc", uint(pid)).Rows()
	if err != nil {
		seelog.Error("[ListCommentByPostID]get data err", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var comment Comment
		DB.ScanRows(rows, &comment)
		//fmt.Println(rows)
		comments = append(comments, &comment)
	}
	return comments, err
}

func CountComment() int {
	var count int
	DB.Model(&Comment{}).Count(&count)
	return count
}
