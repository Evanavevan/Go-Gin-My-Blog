package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/snluu/uuid"
	"path/filepath"
	"github.com/cihub/seelog"
	"fmt"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	"net/http"
	"blog/system"
)

const (
	SessionKey         = "UserID"       // session key
	ContextUserKey     = "User"         // context user key
	SessionGithubState = "GITHUB_STATE" // github state session key
	SessionCaptcha     = "GIN_CAPTCHA"  // captcha session key
)

// 计算字符串的md5值
func Md5(source string) string {
	md5h := md5.New()
	md5h.Write([]byte(source))
	return hex.EncodeToString(md5h.Sum(nil))
}

func Truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n])
	}
	return s
}

func UUID() string {
	return uuid.Rand().Hex()
}

func GetCurrentTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc)
}

func SendToMail(user, password, host, to, subject, body, mailType string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var contentType string
	if mailType == "html" {
		contentType = "Content-Type: text/" + mailType + "; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain" + "; charset=UTF-8"
	}
	msg := []byte("To: " + to + "\r\nFrom: " + user + "\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	sendTo := strings.Split(to, ";")
	return smtp.SendMail(host, auth, user, sendTo, msg)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Decrypt(cipherText []byte, keyString string) ([]byte, error) {
	// Key
	key := []byte(keyString)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Before even testing the decryption,
	// if the text is too small, then it is incorrect
	if len(cipherText) < aes.BlockSize {
		err = errors.New("Text is too short")
		return nil, nil
	}

	// Get the 16 byte IV
	iv := cipherText[:aes.BlockSize]

	// Remove the IV from the cipherText
	cipherText = cipherText[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from cipherText
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}

func Encrypt(plaintext []byte, keyString string) ([]byte, error) {

	// Key
	key := []byte(keyString)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Empty array of 16 + plaintext length
	// Include the IV at the beginning
	cipherText := make([]byte, aes.BlockSize+len(plaintext))

	// Slice of first 16 bytes
	iv := cipherText[:aes.BlockSize]

	// Write 16 rand bytes to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt bytes from plaintext to cipherText
	stream.XORKeyStream(cipherText[aes.BlockSize:], plaintext)

	return cipherText, nil
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		seelog.Critical(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func Handle404(c *gin.Context) {
	HandleMessage(c, http.StatusNotFound, "Sorry, I lost myself!")
}

func HandleMessage(c *gin.Context, code int, message string) {
	c.HTML(code, "errors/error.html", gin.H{
		"message": message,
	})
}

func HtmlSuccess(c *gin.Context, template string, data interface{})  {
	c.HTML(http.StatusOK, template, data)
}


func WriteJSON(ctx *gin.Context, h gin.H) {
	if _, ok := h["succeed"]; !ok {
		h["succeed"] = false
	}
	ctx.JSON(http.StatusOK, h)
}

func SendEmail(to, subject, body string) error {
	c := system.GetConfiguration()
	return SendToMail(c.SmtpUsername, c.SmtpPassword, c.SmtpHost, to, subject, body, "html")
}

func NotifyEmail(subject, body string) error {
	notifyEmailsStr := system.GetConfiguration().NotifyEmails
	if notifyEmailsStr != "" {
		notifyEmails := strings.Split(notifyEmailsStr, ";")
		emails := make([]string, 0)
		for _, email := range notifyEmails {
			if email != "" {
				emails = append(emails, email)
			}
		}
		if len(emails) > 0 {
			return SendEmail(strings.Join(emails, ";"), subject, body)
		}
	}
	return nil
}

func ParseIdToUint(id string, funcName string) (uint64, error) {
	pid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		seelog.Error(fmt.Sprintf("[%s]parse unit err ", funcName), err)
		return 0, err
	}
	return pid, nil
}

func GetUserId(c *gin.Context) uint {
	s := sessions.Default(c)
	sessionUserID := s.Get(SessionKey)
	userId, _ := sessionUserID.(uint)
	return userId
}
