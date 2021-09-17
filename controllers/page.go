package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"blog/models"
	"github.com/cihub/seelog"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"blog/forms"
	. "blog/helpers"
)

func PageGet(c *gin.Context) {
	page, err := models.GetPageById(c.Param("id"))
	if err != nil || !page.IsPublished {
		seelog.Error("[PageGet]get page by id err", err)
		Handle404(c)
		return
	}
	page.View++
	page.UpdateView()
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "page/display.html", gin.H{
		"page": page,
		"user": user,
	})
}

func PageNew(c *gin.Context) {
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "page/new.html", gin.H{
		"user": user,
	})
}

func PageCreate(c *gin.Context) {
	var PageForm forms.PageFrom
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("CheckPublish", forms.CheckPublish)
	}
	if err := c.ShouldBind(&PageForm); err != nil {
		seelog.Error("[PageCreate]input param err", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	published := "on" == PageForm.IsPublished
	page := &models.Page{
		Title:       PageForm.Title,
		Body:        PageForm.Body,
		IsPublished: published,
	}
	err := page.Insert()
	if err != nil {
		seelog.Error("[PageCreate]insert page err", err)
		HtmlSuccess(c, "page/new.html", gin.H{
			"message": err.Error(),
			"page":    page,
		})
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/admin/page")

}

func PageEdit(c *gin.Context) {
	page, err := models.GetPageById(c.Param("id"))
	if err != nil {
		seelog.Error("[PageEdit]get page by id err", err)
		Handle404(c)
	}
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "page/modify.html", gin.H{
		"page": page,
		"user": user,
	})
}

func PageUpdate(c *gin.Context) {
	var PageForm forms.PageFrom
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("CheckPublish", forms.CheckPublish)
	}
	if err := c.ShouldBind(&PageForm); err != nil {
		seelog.Error("[PageUpdate]input param err", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	published := "on" == PageForm.IsPublished
	pid, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		seelog.Error("[PageUpdate]parse uint err", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	page := &models.Page{Title: PageForm.Title, Body: PageForm.Body, IsPublished: published}
	page.ID = uint(pid)
	err = page.Update()
	if err != nil {
		seelog.Error("[PageUpdate]update page err", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/admin/page")
}

func PagePublish(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	page, err := models.GetPageById(c.Param("id"))
	if err == nil {
		seelog.Error("[PagePublish]get page by id err", err)
		res["message"] = err.Error()
		return
	}
	page.IsPublished = !page.IsPublished
	err = page.Update()
	if err == nil {
		seelog.Error("[PagePublish]update page err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func PageDelete(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	pid, err := ParseIdToUint(c.Param("id"), "PageDelete")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	page := &models.Page{}
	page.ID = uint(pid)
	err = page.Delete()
	if err != nil {
		seelog.Error("[PageDelete]delete page err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func PageIndex(c *gin.Context) {
	pages, _ := models.ListAllPage()
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "admin/page.html", gin.H{
		"pages":    pages,
		"user":     user,
		"comments": models.MustListUnreadComment(),
	})
}
