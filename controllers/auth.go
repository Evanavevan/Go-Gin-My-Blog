package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"blog/helpers"
	"blog/system"
	. "blog/helpers"
)

func AuthGet(c *gin.Context) {
	authType := c.Param("authType")

	session := sessions.Default(c)
	uuid := helpers.UUID()
	session.Delete(SessionGithubState)
	session.Set(SessionGithubState, uuid)
	session.Save()

	var authUrl string
	switch authType {
	case "github":
		authUrl = fmt.Sprintf(system.GetConfiguration().GithubAuthUrl, system.GetConfiguration().GithubClientId, uuid)
	case "weibo":
		HandleMessage(c, http.StatusNotImplemented, "Sorry, not implemented!")
	case "qq":
		HandleMessage(c, http.StatusNotImplemented, "Sorry, not implemented!")
	case "wechat":
		HandleMessage(c, http.StatusNotImplemented, "Sorry, not implemented!")
	default:
		authUrl = "/user/login"
	}
	c.Redirect(http.StatusFound, authUrl)
}
