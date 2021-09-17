package forms

type LinkFrom struct {
	Name 	string `form:"name" json:"name" binding:"required"`
	Url 	string `form:"url" json:"url" binding:"required,uri"`
	Sort  	string `form:"sort" json:"sort" binding:"required,number"`
}

