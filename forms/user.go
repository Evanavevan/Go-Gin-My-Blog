package forms

import (
	"regexp"
	"github.com/cihub/seelog"
	"github.com/go-playground/validator/v10"
)

type RegisterFrom struct {
	// 邮箱
	Email string `form:"email" json:"email" binding:"required,email"`
	// 手机号
	Telephone string `form:"telephone" json:"telephone" binding:"required,RegexTelephone"`
	// 密码  binding:"required"为必填字段,长度大于6小于12
	PassWord  string `form:"password" json:"password" binding:"required,min=6,max=12"`
	// 二次密码
	PassWord2  string `form:"password2" json:"password2" binding:"eqfield=PassWord"`
}

type LoginForm struct {
	// 邮箱
	Email string `form:"email" json:"email" binding:"required,email"`
	// 密码  binding:"required"为必填字段,长度大于6小于12
	PassWord  string `form:"password" json:"password" binding:"required,min=6,max=12"`
}

func RegexTelephone(fl validator.FieldLevel) bool {
	regex := regexp.MustCompile("^1[3|4|5|7|8][0-9]{9}$")
	if regex == nil {
		seelog.Error("[RegexTelephone]regexp init err")
	}
	return regex.MatchString(fl.Field().Interface().(string))
}
