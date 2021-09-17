package controllers

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/cihub/seelog"
	. "blog/helpers"
)

func Upload(c *gin.Context) {
	var (
		err      error
		res      = gin.H{}
		url      string
		uploader Uploader
		file     multipart.File
		fh       *multipart.FileHeader
	)
	defer WriteJSON(c, res)
	file, fh, err = c.Request.FormFile("file")
	if err != nil {
		seelog.Error("[Upload]get file err", err)
		res["message"] = err.Error()
		return
	}

	//uploader = QiniuUploader{}
	uploader = SmmsUploader{}

	url, err = uploader.upload(file, fh)
	if err != nil {
		seelog.Error("[Upload]upload file err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
	res["url"] = url
}
