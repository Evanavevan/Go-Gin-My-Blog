package forms

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

type SubscribeForm struct {
	// 邮箱
	Email string `form:"email" json:"email" binding:"required,email"`
}

type SubscriberForm struct {
	// 邮箱
	Email string `form:"email" json:"email" binding:"required,CheckEmail"`
	Subject string `form:"subject" json:"subject" binding:"required"`
	Body string `form:"body" json:"body" binding:"required"`
}

func CheckEmail(fl validator.FieldLevel) bool {
	var res bool
	email := fl.Field().Interface().(string)
	if email == "" {
		res = true
	} else {
		//匹配电子邮箱
		pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
		reg := regexp.MustCompile(pattern)
		res = reg.MatchString(email)
	}
	return res
}