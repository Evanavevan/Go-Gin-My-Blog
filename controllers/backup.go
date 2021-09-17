package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	. "blog/helpers"
	"blog/system"
)

func BackupPost(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	err = Backup()
	if err != nil {
		seelog.Error("[BackupPost]backup err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func RestorePost(c *gin.Context) {
	var (
		fileName  string
		fileUrl   string
		err       error
		res       = gin.H{}
		resp      *http.Response
		bodyBytes []byte
	)
	defer WriteJSON(c, res)
	fileName = c.PostForm("fileName")
	if fileName == "" {
		res["message"] = "fileName cannot be empty."
		return
	}
	fileUrl = system.GetConfiguration().QiniuFileServer + fileName
	resp, err = http.Get(fileUrl)
	if err != nil {
		seelog.Error("[RestorePost]get file err", err)
		res["message"] = err.Error()
		return
	}
	defer resp.Body.Close()

	bodyBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		seelog.Error("[RestorePost]read body err", err)
		res["message"] = err.Error()
		return
	}
	bodyBytes, err = Decrypt(bodyBytes, system.GetConfiguration().BackupKey)
	if err != nil {
		seelog.Error("[RestorePost]decrypt body err", err)
		res["message"] = err.Error()
		return
	}
	err = ioutil.WriteFile(fileName, bodyBytes, os.ModePerm)
	if err == nil {
		seelog.Error("[RestorePost]write file err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func Backup() (err error) {
	var (
		u           *url.URL
		exist       bool
		ret         PutRet
		bodyBytes   []byte
		encryptData []byte
	)
	u, err = url.Parse(system.GetConfiguration().DSN)
	if err != nil {
		seelog.Debug("[Backup]parse dsn error:%v", err)
		return
	}
	exist, _ = PathExists(u.Path)
	if !exist {
		err = errors.New("database file doesn't exists.")
		seelog.Debug("[Backup]database file doesn't exists.")
		return
	}
	seelog.Debug("start backup...")
	bodyBytes, err = ioutil.ReadFile(u.Path)
	if err != nil {
		seelog.Error("[Backup]read file err", err)
		seelog.Error(err)
		return
	}
	encryptData, err = Encrypt(bodyBytes, system.GetConfiguration().BackupKey)
	if err != nil {
		seelog.Error("[Backup]encrypt file err", err)
		seelog.Error(err)
		return
	}

	putPolicy := storage.PutPolicy{
		Scope: system.GetConfiguration().QiniuBucket,
	}
	mac := qbox.NewMac(system.GetConfiguration().QiniuAccessKey, system.GetConfiguration().QiniuSecretKey)
	token := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	uploader := storage.NewFormUploader(&cfg)
	putExtra := storage.PutExtra{}

	fileName := fmt.Sprintf("wblog_%s.db", GetCurrentTime().Format("20060102150405"))
	err = uploader.Put(context.Background(), &ret, token, fileName, bytes.NewReader(encryptData), int64(len(encryptData)), &putExtra)
	if err != nil {
		seelog.Debugf("backup error:%v", err)
		return
	}
	seelog.Debug("backup successfully.")
	return err
}
