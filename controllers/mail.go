package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"blog/models"
	. "blog/helpers"
	"github.com/cihub/seelog"
)

func SendMail(c *gin.Context) {
	var (
		err        error
		res        = gin.H{}
		uid        uint64
		subscriber *models.Subscriber
	)
	defer WriteJSON(c, res)
	subject := c.PostForm("subject")
	content := c.PostForm("content")

	if subject == "" || content == "" {
		res["message"] = "error parameter"
		return
	}
	uid, err = ParseIdToUint(c.Query("userId"), "SendMail")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	subscriber, err = models.GetSubscriberById(uint(uid))
	if err != nil {
		seelog.Error("[SendMail]get subscriber by id err", err)
		res["message"] = err.Error()
		return
	}
	err = SendEmail(subscriber.Email, subject, content)
	if err != nil {
		seelog.Error("[SendMail]send email err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func SendBatchMail(c *gin.Context) {
	var (
		err         error
		res         = gin.H{}
		subscribers []*models.Subscriber
		emails      = make([]string, 0)
	)
	defer WriteJSON(c, res)
	subject := c.PostForm("subject")
	content := c.PostForm("content")
	if subject == "" || content == "" {
		res["message"] = "error parameter"
		return
	}
	subscribers, err = models.ListSubscriber(true)
	if err != nil {
		seelog.Error("[SendBatchMail]list subscriber err", err)
		res["message"] = err.Error()
		return
	}
	for _, subscriber := range subscribers {
		emails = append(emails, subscriber.Email)
	}
	err = SendEmail(strings.Join(emails, ";"), subject, content)
	if err != nil {
		seelog.Error("[SendBatchMail]send email err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
