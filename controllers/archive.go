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
	"github.com/cihub/seelog"
	. "blog/helpers"
)

func ArchiveGet(c *gin.Context) {
	var (
		year      string
		month     string
		page      string
		pageIndex int
		pageSize  = system.GetConfiguration().PageSize
		total     int
		err       error
		posts     []*models.Post
		policy    *bluemonday.Policy
	)
	year = c.Param("year")
	month = c.Param("month")
	page = c.Query("page")
	pageIndex, _ = strconv.Atoi(page)
	if pageIndex <= 0 {
		pageIndex = 1
	}
	posts, err = models.ListPostByArchive(year, month, pageIndex, pageSize)
	if err != nil {
		seelog.Info("[ArchiveGet]list archive err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	total, err = models.CountPostByArchive(year, month)
	if err != nil {
		seelog.Info("[ArchiveGet]count archive err", err)
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
		"user": user,
	})

}
