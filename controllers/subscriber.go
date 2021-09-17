package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	. "blog/helpers"
	"blog/models"
	"blog/system"
	"github.com/cihub/seelog"
	"blog/forms"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func SubscribeGet(c *gin.Context) {
	count, _ := models.CountSubscriber()
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "other/subscribe.html", gin.H{
		"user": user,
		"total": count,
	})
}

func Subscribe(c *gin.Context) {
	var (
		err error
		count int
		mail string
		SubscribeForm forms.SubscribeForm
	)
	if e := c.ShouldBind(&SubscribeForm); e != nil {
		seelog.Error("[Subscribe]validate err", e)
		err = errors.New("input params error")
		goto response
	}
	mail = SubscribeForm.Email
	if len(mail) > 0 {
		var subscriber *models.Subscriber
		subscriber, err = models.GetSubscriberByEmail(mail)
		if err == nil {
			if !subscriber.VerifyState && GetCurrentTime().After(subscriber.OutTime) { //激活链接超时
				err = sendActiveEmail(subscriber)
				if err == nil {
					count, _ = models.CountSubscriber()
					err = errors.New("subscribe succeed")
					goto response
				}
			} else if subscriber.VerifyState && !subscriber.SubscribeState { //已认证，未订阅
				subscriber.SubscribeState = true
				err = subscriber.Update()
				if err == nil {
					err = errors.New("subscribe succeed.")
				}
			} else {
				err = errors.New("mail have already been active or have inactive mail in your mailbox.")
			}
		} else {
			subscriber := &models.Subscriber{
				Email: mail,
			}
			err = subscriber.Insert()
			if err == nil {
				err = sendActiveEmail(subscriber)
				if err == nil {
					count, _ = models.CountSubscriber()
					err = errors.New("subscribe succeed")
					goto response
				}
			}
		}
	} else {
		err = errors.New("empty mail address.")
	}
	count, _ = models.CountSubscriber()
	response:
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "other/subscribe.html", gin.H{
		"message": err.Error(),
		"total":   count,
		"user": user,
	})
}

func sendActiveEmail(subscriber *models.Subscriber) (err error) {
	uuid := UUID()
	duration, _ := time.ParseDuration("30m")
	subscriber.OutTime = GetCurrentTime().Add(duration)
	subscriber.SecretKey = uuid
	signature := Md5(subscriber.Email + uuid + subscriber.OutTime.Format("20060102150405"))
	subscriber.Signature = signature
	err = SendEmail(subscriber.Email, "[blog]邮箱验证", fmt.Sprintf("%s/active?sid=%s", system.GetConfiguration().Domain, signature))
	if err != nil {
		seelog.Error("[sendActiveEmail]send active email err ", err)
		return
	}
	err = subscriber.Update()
	return
}

func ActiveSubscriber(c *gin.Context) {
	var (
		err        error
		subscriber *models.Subscriber
	)
	sid := c.Query("sid")
	if sid == "" {
		HandleMessage(c, http.StatusBadRequest, "激活链接有误，请重新获取！")
		return
	}
	subscriber, err = models.GetSubscriberBySignature(sid)
	if err != nil {
		seelog.Error("[ActiveSubscriber]get subscriber by signature err", err)
		HandleMessage(c, http.StatusBadRequest,"激活链接有误，请重新获取！")
		return
	}
	if !GetCurrentTime().Before(subscriber.OutTime) {
		HandleMessage(c, http.StatusBadRequest, "激活链接已过期，请重新获取！")
		return
	}
	subscriber.VerifyState = true
	subscriber.OutTime = GetCurrentTime()
	err = subscriber.Update()
	if err != nil {
		HandleMessage(c, http.StatusBadRequest, fmt.Sprintf("激活失败！%s", err.Error()))
		return
	}
	HandleMessage(c, http.StatusBadRequest, "激活成功！")
}

func UnSubscribe(c *gin.Context) {
	sid := c.Query("sid")
	if sid == "" {
		HandleMessage(c, http.StatusBadRequest, "Internal Server Error!")
		return
	}
	subscriber, err := models.GetSubscriberBySignature(sid)
	if err != nil || !subscriber.VerifyState || !subscriber.SubscribeState {
		HandleMessage(c, http.StatusBadRequest, "Unsubscribe failed.")
		return
	}
	subscriber.SubscribeState = false
	err = subscriber.Update()
	if err == nil {
		HandleMessage(c, http.StatusBadRequest, fmt.Sprintf("Unscribe failed.%s", err.Error()))
		return
	}
	HandleMessage(c, http.StatusBadRequest, "Unsubscribe Successful!")
}

func GetUnSubcribeUrl(subscriber *models.Subscriber) (string, error) {
	uuid := UUID()
	signature := Md5(subscriber.Email + uuid)
	subscriber.SecretKey = uuid
	subscriber.Signature = signature
	err := subscriber.Update()
	return fmt.Sprintf("%s/unsubscribe?sid=%s", system.GetConfiguration().Domain, signature), err
}

func sendEmailToSubscribers(subject, body string) (err error) {
	var (
		subscribers []*models.Subscriber
		emails      = make([]string, 0)
	)
	subscribers, err = models.ListSubscriber(true)
	if err != nil {
		seelog.Error("[sendEmailToSubscribers]list subscriber err", err)
		return
	}
	for _, subscriber := range subscribers {
		emails = append(emails, subscriber.Email)
	}
	if len(emails) == 0 {
		err = errors.New("no subscribers!")
		return
	}
	err = SendEmail(strings.Join(emails, ";"), subject, body)
	return
}

func SubscriberIndex(c *gin.Context) {
	subscribers, _ := models.ListSubscriber(false)
	user, _ := c.Get(ContextUserKey)
	c.HTML(http.StatusOK, "admin/subscriber.html", gin.H{
		"subscribers": subscribers,
		"user":        user,
		"comments":    models.MustListUnreadComment(),
	})
}

// 邮箱为空时，发送给所有订阅者
func SubscriberPost(c *gin.Context) {
	var (
		err error
		res = gin.H{}
		SubscriberForm forms.SubscriberForm
	)
	defer WriteJSON(c, res)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("CheckEmail", forms.CheckEmail)
	}
	if e := c.ShouldBind(&SubscriberForm); e != nil {
		seelog.Error("[SubscriberPost]input param err", e)
		res["message"] = e.Error()
		return
	}
	mail := SubscriberForm.Email
	subject := SubscriberForm.Subject
	body := SubscriberForm.Body
	if len(mail) > 0 {
		err = SendEmail(mail, subject, body)
	} else {
		err = sendEmailToSubscribers(subject, body)
	}
	if err != nil {
		seelog.Error("[SubscriberPost]send email fail", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
