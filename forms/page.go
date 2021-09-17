package forms

import "github.com/go-playground/validator/v10"

type PageFrom struct {
	Title 	string `form:"title" json:"title" binding:"required"`
	Body 	string `form:"body" json:"body" binding:"required"`
	IsPublished  	string `form:"isPublished" json:"isPublished" binding:"required,CheckPublish"`
}

func CheckPublish(fl validator.FieldLevel) bool {
	var res bool
	switch fl.Field().Interface().(string) {
	case "on":
	case "off":
		res = true
		break
	default:
		res = false
		break
	}
	return res
}