package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"blog/models"
	"github.com/cihub/seelog"
	"blog/forms"
	. "blog/helpers"
)

func LinkIndex(c *gin.Context) {
	links, _ := models.ListLinks()
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "admin/link.html", gin.H{
		"links":    links,
		"user":     user,
		"comments": models.MustListUnreadComment(),
	})
}

func LinkCreate(c *gin.Context) {
	var (
		err   error
		res   = gin.H{}
		sort uint64
		LinkForm forms.LinkFrom
	)
	defer WriteJSON(c, res)
	if e := c.ShouldBind(&LinkForm); e != nil {
		seelog.Error("[LinkCreate]input param err", e)
		res["message"] = "input param err"
		return
	}
	sort, err = ParseIdToUint(LinkForm.Sort, "LinkCreate")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	link := &models.Link{
		Name: LinkForm.Name,
		Url:  LinkForm.Url,
		Sort: int(sort),
	}
	err = link.Insert()
	if err != nil {
		seelog.Error("[LinkCreate]insert link err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func LinkUpdate(c *gin.Context) {
	var (
		id   uint64
		sort uint64
		err   error
		res   = gin.H{}
		LinkForm forms.LinkFrom
	)
	defer WriteJSON(c, res)
	if e := c.ShouldBind(&LinkForm); e != nil {
		seelog.Error("[LinkUpdate]input param err", e)
		res["message"] = "input param err"
		return
	}
	id, err = ParseIdToUint(c.Param("id"), "LinkUpdate")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	sort, err = ParseIdToUint(LinkForm.Sort, "LinkUpdate")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	link := &models.Link{
		Name: LinkForm.Name,
		Url:  LinkForm.Url,
		Sort: int(sort),
	}
	link.ID = uint(id)
	err = link.Update()
	if err != nil {
		seelog.Error("[LinkUpdate]update link err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func LinkGet(c *gin.Context) {
	id, _ := ParseIdToUint(c.Param("id"), "LinkGet")
	link, err := models.GetLinkById(uint(id))
	if err != nil {
		seelog.Error("[LinkGet]get link by id err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	link.View++
	link.Update()
	c.Redirect(http.StatusFound, link.Url)
}

func LinkDelete(c *gin.Context) {
	var (
		err error
		id uint64
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	id, err = ParseIdToUint(c.Param("id"), "LinkDelete")
	if err != nil {
		res["message"] = err.Error()
		return
	}

	link := new(models.Link)
	link.ID = uint(id)
	err = link.Delete()
	if err != nil {
		seelog.Error("[LinkDelete]delete link err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
