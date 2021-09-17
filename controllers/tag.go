package controllers

import (
	"net/http"
	"strconv"

	"math"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"blog/models"
	"blog/system"
	. "blog/helpers"
	"github.com/cihub/seelog"
)

func TagCreate(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	tag := &models.Tag{Name: c.PostForm("name")}
	err = tag.Insert()
	if err != nil {
		seelog.Error("[TagCreate]insert tag err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func TagUpdate(c *gin.Context) {
	var (
		id   uint64
		err   error
		res   = gin.H{}
	)
	defer WriteJSON(c, res)
	name := c.PostForm("name")
	if len(name) == 0 {
		res["message"] = "error parameter"
		return
	}
	id, err = ParseIdToUint(c.Param("id"), "TagUpdate")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	tag := &models.Tag{
		Name: name,
	}
	tag.ID = uint(id)
	err = tag.Update()
	if err != nil {
		seelog.Error("[TagUpdate]update tag err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func TagDelete(c *gin.Context) {
	var (
		err error
		id uint64
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	id, err = ParseIdToUint(c.Param("id"), "TagDelete")
	if err != nil {
		res["message"] = err.Error()
		return
	}

	tag := new(models.Tag)
	tag.ID = uint(id)
	err = tag.Delete()
	if err != nil {
		seelog.Error("[TagDelete]delete tag err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func TagGet(c *gin.Context) {
	var (
		tagName   string
		page      string
		pageIndex int
		pageSize  = system.GetConfiguration().PageSize
		total     int
		err       error
		policy    *bluemonday.Policy
		posts     []*models.Post
	)
	tagName = c.Param("tag")
	page = c.Query("page")
	pageIndex, _ = strconv.Atoi(page)
	if pageIndex <= 0 {
		pageIndex = 1
	}
	posts, err = models.ListPublishedPost(tagName, pageIndex, pageSize)
	if err != nil {
		seelog.Error("[TagGet]list publish post err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	total, err = models.CountPostByTag(tagName)
	if err != nil {
		seelog.Error("[TagGet]count post by tag err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	policy = bluemonday.StrictPolicy()
	for _, post := range posts {
		post.Tags, _ = models.ListTagByPostId(strconv.FormatUint(uint64(post.ID), 10))
		post.Body = policy.Sanitize(string(blackfriday.Run([]byte(post.Body), blackfriday.WithNoExtensions())))
	}
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "index/index.html", gin.H{
		"posts":           posts,
		"tags":            models.MustListTag(),
		"archives":        models.MustListPostArchives(),
		"links":           models.MustListLinks(),
		"pageIndex":       pageIndex,
		"totalPage":       int(math.Ceil(float64(total) / float64(pageSize))),
		"maxReadPosts":    models.MustListMaxReadPost(),
		"maxCommentPosts": models.MustListMaxCommentPost(),
		"user": 		   user,
	})
}

func TagIndex(c *gin.Context) {
	tags, _ := models.ListAllTag()
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "admin/tag.html", gin.H{
		"tags":    tags,
		"user":     user,
		"comments": models.MustListUnreadComment(),
	})
}
