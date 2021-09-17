package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"blog/models"
	"github.com/cihub/seelog"
	"blog/forms"
	. "blog/helpers"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func PostGet(c *gin.Context) {
	id := c.Param("id")
	post, err := models.GetPostById(id)
	if err != nil || !post.IsPublished {
		seelog.Error("[PostGet]get post by id err", err)
		Handle404(c)
		return
	}
	post.View++
	post.UpdateView()
	post.Tags, _ = models.ListTagByPostId(id)
	post.Comments, _ = models.ListCommentByPostID(id)
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "post/display.html", gin.H{
		"post": post,
		"user": user,
	})
}

func PostNew(c *gin.Context) {
	tags, _ := models.ListAllTag()
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "post/new.html", gin.H{
		"user": user,
		"tags": tags,
	})
}

func PostCreate(c *gin.Context) {
	var PageForm forms.PageFrom
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("CheckPublish", forms.CheckPublish)
	}
	if err := c.ShouldBind(&PageForm); err != nil {
		seelog.Error("[PostCreate]input param err", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	tags := c.PostForm("tags")
	isPublished := PageForm.IsPublished
	published := "on" == isPublished

	post := &models.Post{
		Title:       PageForm.Title,
		Body:        PageForm.Body,
		IsPublished: published,
	}
	err := post.Insert()
	if err != nil {
		seelog.Error("[PostCreate]insert post err", err)
		user, _ := c.Get(ContextUserKey)
		HtmlSuccess(c, "post/new.html", gin.H{
			"post":    post,
			"message": err.Error(),
			"user": user,
		})
		return
	}

	// add tag for post
	if len(tags) > 0 {
		tagArr := strings.Split(tags, ",")
		for _, tag := range tagArr {
			tagId, err := ParseIdToUint(tag, "PostCreate")
			if err != nil {
				continue
			}
			pt := &models.PostTag{
				PostId: post.ID,
				TagId:  uint(tagId),
			}
			pt.Insert()
		}
	}
	c.Redirect(http.StatusMovedPermanently, "/admin/post")
}

func PostEdit(c *gin.Context) {
	post, err := models.GetPostById(c.Param("id"))
	if err != nil {
		seelog.Error("[PostEdit]get post by id err", err)
		Handle404(c)
		return
	}
	tags, _ := models.ListAllTag()
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "post/modify.html", gin.H{
		"post": post,
		"tags": tags,
		"user": user,
	})
}

func PostUpdate(c *gin.Context) {
	var PageForm forms.PageFrom
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("CheckPublish", forms.CheckPublish)
	}
	if err := c.ShouldBind(&PageForm); err != nil {
		seelog.Error("[PostCreate]input param err", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	tags := c.PostForm("tags")
	published := "on" == PageForm.IsPublished

	pid, err := ParseIdToUint(c.Param("id"), "PostUpdate")
	if err != nil {
		Handle404(c)
		return
	}

	post := &models.Post{
		Title:       PageForm.Title,
		Body:        PageForm.Body,
		IsPublished: published,
	}
	post.ID = uint(pid)
	err = post.Update()
	if err != nil {
		seelog.Error("[PostUpdate]update post err", err)
		user, _ := c.Get(ContextUserKey)
		HtmlSuccess(c, "post/modify.html", gin.H{
			"post":    post,
			"message": err.Error(),
			"user": user,
		})
		return
	}
	// 删除tag
	models.DeletePostTagByPostId(post.ID)
	// 添加tag
	if len(tags) > 0 {
		tagArr := strings.Split(tags, ",")
		for _, tag := range tagArr {
			tagId, err := ParseIdToUint(tag, "PostUpdate")
			if err != nil {
				continue
			}
			pt := &models.PostTag{
				PostId: post.ID,
				TagId:  uint(tagId),
			}
			pt.Insert()
		}
	}
	c.Redirect(http.StatusMovedPermanently, "/admin/post")
}

func PostPublish(c *gin.Context) {
	var (
		err  error
		res  = gin.H{}
		post *models.Post
	)
	defer WriteJSON(c, res)
	post, err = models.GetPostById(c.Param("id"))
	if err != nil {
		seelog.Error("[PostPublish]get post by id err", err)
		res["message"] = err.Error()
		return
	}
	post.IsPublished = !post.IsPublished
	err = post.Update()
	if err != nil {
		seelog.Error("[PostPublish]update post err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func PostDelete(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	pid, err := ParseIdToUint(c.Param("id"), "PostDelete")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	post := &models.Post{}
	post.ID = uint(pid)
	err = post.Delete()
	if err != nil {
		seelog.Error("[PostDelete]delete post err", err)
		res["message"] = err.Error()
		return
	}
	models.DeletePostTagByPostId(uint(pid))
	res["succeed"] = true
}

func PostIndex(c *gin.Context) {
	posts, _ := models.ListAllPost("")
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "admin/post.html", gin.H{
		"posts":    posts,
		"user":     user,
		"comments": models.MustListUnreadComment(),
	})
}
