package controllers

import (
	"fmt"

	"github.com/dchest/captcha"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"blog/models"
	"blog/system"
	. "blog/helpers"
	"github.com/cihub/seelog"
)

func CommentPost(c *gin.Context) {
	var (
		err  error
		res  = gin.H{}
		post *models.Post
	)
	defer WriteJSON(c, res)
	s := sessions.Default(c)
	userId := GetUserId(c)

	verifyCode := c.PostForm("verifyCode")
	captchaId := s.Get(SessionCaptcha)
	s.Delete(SessionCaptcha)
	_captchaId, _ := captchaId.(string)
	if !captcha.VerifyString(_captchaId, verifyCode) {
		res["message"] = "error verifycode"
		return
	}

	postId := c.PostForm("postId")
	content := c.PostForm("content")
	if len(content) == 0 {
		res["message"] = "content cannot be empty."
		return
	}

	post, err = models.GetPostById(postId)
	if err != nil {
		seelog.Error("[CommentPost]get post id err", err)
		res["message"] = err.Error()
		return
	}
	pid, err := ParseIdToUint(postId, "CommentPost")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	comment := &models.Comment{
		PostID:  uint(pid),
		Content: content,
		UserID:  userId,
	}
	err = comment.Insert()
	if err != nil {
		seelog.Error("[CommentPost]insert comment err", err)
		res["message"] = err.Error()
		return
	}
	NotifyEmail("[blog]您有一条新评论", fmt.Sprintf("<a href=\"%s/post/%d\" target=\"_blank\">%s</a>:%s", system.GetConfiguration().Domain, post.ID, post.Title, content))
	res["succeed"] = true
}

func CommentDelete(c *gin.Context) {
	var (
		err error
		res = gin.H{}
		cid uint64
	)
	defer WriteJSON(c, res)

	userId := GetUserId(c)

	cid, err = ParseIdToUint(c.Param("id"), "CommentDelete")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	comment := &models.Comment{
		UserID: uint(userId),
	}
	comment.ID = uint(cid)
	err = comment.Delete()
	if err != nil {
		seelog.Error("[CommentDelete]delete comment err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func CommentRead(c *gin.Context) {
	var (
		id uint64
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	id, err = ParseIdToUint(c.Param("id"), "CommentRead")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	comment := new(models.Comment)
	comment.ID = uint(id)
	err = comment.Update()
	if err != nil {
		seelog.Error("[CommentRead]update comment err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func CommentReadAll(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	err = models.SetAllCommentRead()
	if err != nil {
		seelog.Error("[CommentReadAll]update all comment err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
