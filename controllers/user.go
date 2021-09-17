package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/alimoeeny/gooauth2"
	"github.com/cihub/seelog"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	. "blog/helpers"
	"blog/models"
	"blog/system"
	"blog/forms"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type GithubUserInfo struct {
	AvatarURL         string      `json:"avatar_url"`
	Bio               interface{} `json:"bio"`
	Blog              string      `json:"blog"`
	Company           interface{} `json:"company"`
	CreatedAt         string      `json:"created_at"`
	Email             interface{} `json:"email"`
	EventsURL         string      `json:"events_url"`
	Followers         int         `json:"followers"`
	FollowersURL      string      `json:"followers_url"`
	Following         int         `json:"following"`
	FollowingURL      string      `json:"following_url"`
	GistsURL          string      `json:"gists_url"`
	GravatarID        string      `json:"gravatar_id"`
	Hireable          interface{} `json:"hireable"`
	HTMLURL           string      `json:"html_url"`
	ID                int         `json:"id"`
	Location          interface{} `json:"location"`
	Login             string      `json:"login"`
	Name              interface{} `json:"name"`
	OrganizationsURL  string      `json:"organizations_url"`
	PublicGists       int         `json:"public_gists"`
	PublicRepos       int         `json:"public_repos"`
	ReceivedEventsURL string      `json:"received_events_url"`
	ReposURL          string      `json:"repos_url"`
	SiteAdmin         bool        `json:"site_admin"`
	StarredURL        string      `json:"starred_url"`
	SubscriptionsURL  string      `json:"subscriptions_url"`
	Type              string      `json:"type"`
	UpdatedAt         string      `json:"updated_at"`
	URL               string      `json:"url"`
}

func LoginGet(c *gin.Context) {
	HtmlSuccess(c, "auth/login.html", nil)
}

func RegisterGet(c *gin.Context) {
	HtmlSuccess(c, "auth/register.html", nil)
}

func LogoutGet(c *gin.Context) {
	s := sessions.Default(c)
	s.Clear()
	s.Save()
	c.Redirect(http.StatusSeeOther, "/user/login")
}

func RegisterPost(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("RegexTelephone", forms.RegexTelephone)
	}
	var RegisterFrom forms.RegisterFrom
	if e := c.ShouldBind(&RegisterFrom); e != nil {
		seelog.Error("[RegisterPost]validate err", e)
		res["message"] = "input params error"
		return
	}
	user := &models.User{
		Email:     RegisterFrom.Email,
		Telephone: RegisterFrom.Telephone,
		Password:  RegisterFrom.PassWord,
		// todo
		IsAdmin:   true,
	}
	user.Password = Md5(user.Email + user.Password)
	err = user.Insert()
	if err != nil {
		seelog.Error("[RegisterPost]insert user err", err)
		res["message"] = "email already exists"
		return
	}
	res["succeed"] = true
}

func LoginPost(c *gin.Context) {
	var (
		err  error
		user *models.User
	)
	var LoginFrom forms.LoginForm
	if e := c.ShouldBind(&LoginFrom); e != nil {
		seelog.Error("[LoginPost]validate err", e)
		HtmlSuccess(c, "auth/login.html", gin.H{
			"message": "input params error",
		})
		return
	}
	user, err = models.GetUserByUsername(LoginFrom.Email)
	if err != nil || user.Password != Md5(LoginFrom.Email + LoginFrom.PassWord) {
		HtmlSuccess(c, "auth/login.html", gin.H{
			"message": "invalid username or password",
		})
		return
	}
	if user.LockState {
		HtmlSuccess(c, "auth/login.html", gin.H{
			"message": "Your account have been locked",
		})
		return
	}
	s := sessions.Default(c)
	s.Clear()
	s.Set(SessionKey, user.ID)
	s.Save()
	if user.IsAdmin {
		c.Redirect(http.StatusMovedPermanently, "/admin/index")
	} else {
		c.Redirect(http.StatusMovedPermanently, "/")
	}
}

func Oauth2Callback(c *gin.Context) {
	var (
		userInfo *GithubUserInfo
		user     *models.User
	)
	code := c.Query("code")
	state := c.Query("state")

	// validate state
	session := sessions.Default(c)
	if len(state) == 0 || (state != session.Get(SessionGithubState) && session.Get(SessionGithubState) != nil) {
		seelog.Errorf("[Oauth2Callback]state not equal: %s != %s", state, session.Get(SessionGithubState))
		c.Abort()
		return
	}
	// remove state from session
	session.Delete(SessionGithubState)
	session.Save()

	// exchange accesstoken by code
	token, err := exchangeTokenByCode(code)
	if err != nil {
		seelog.Error("[Oauth2Callback]change token by code err", err)
		c.Redirect(http.StatusMovedPermanently, "/user/login")
		return
	}

	//get github userinfo by accesstoken
	userInfo, err = getGithubUserInfoByAccessToken(token)
	//fmt.Println(userInfo)
	if err != nil {
		seelog.Error("[Oauth2Callback]get github user info by token err", err)
		c.Redirect(http.StatusMovedPermanently, "/user/login")
		return
	}

	sessionUser, exists := c.Get(ContextUserKey)
	if exists { // 已登录
		user, _ = sessionUser.(*models.User)
		_, err1 := models.IsGithubIdExists(userInfo.Login, user.ID)
		if err1 != nil { // 未绑定
			if user.IsAdmin {
				user.GithubLoginId = userInfo.Login
			}
			user.AvatarUrl = userInfo.AvatarURL
			user.GithubUrl = userInfo.HTMLURL
			err = user.UpdateGithubUserInfo()
		} else {
			err = errors.New("this github loginId has bound another account.")
		}
	} else {
		user = &models.User{
			GithubLoginId: userInfo.Login,
			AvatarUrl:     userInfo.AvatarURL,
			GithubUrl:     userInfo.HTMLURL,
		}
		user, err = user.FirstOrCreate()
		if err == nil {
			if user.LockState {
				err = errors.New("Your account have been locked.")
				HandleMessage(c, http.StatusBadRequest, "Your account have been locked.")
				return
			}
		}
	}

	if err == nil {
		s := sessions.Default(c)
		s.Clear()
		s.Set(SessionKey, user.ID)
		s.Save()
		if user.IsAdmin {
			c.Redirect(http.StatusMovedPermanently, "/admin/index")
		} else {
			c.Redirect(http.StatusMovedPermanently, "/")
		}
	}
}

func exchangeTokenByCode(code string) (accessToken string, err error) {
	var (
		transport *oauth.Transport
		token     *oauth.Token
	)
	transport = &oauth.Transport{Config: &oauth.Config{
		ClientId:     system.GetConfiguration().GithubClientId,
		ClientSecret: system.GetConfiguration().GithubClientSecret,
		RedirectURL:  system.GetConfiguration().GithubRedirectURL,
		TokenURL:     system.GetConfiguration().GithubTokenUrl,
		Scope:        system.GetConfiguration().GithubScope,
	}}
	token, err = transport.Exchange(code)
	if err != nil {
		seelog.Error("[exchangeTokenByCode]change token err", err)
		return
	}
	accessToken = token.AccessToken
	// cache token
	tokenCache := oauth.CacheFile("./blog/vendor/request.token")
	if err := tokenCache.PutToken(token); err != nil {
		seelog.Error("[exchangeTokenByCode]put token err", err)
	}
	return
}

func getGithubUserInfoByAccessToken(token string) (*GithubUserInfo, error) {
	var (
		resp *http.Response
		body []byte
		err  error
	)
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/user?access_token=%s", token), nil)
	// 请求头不可缺
	req.Header.Set("Authorization", "token " + token)
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		seelog.Error("[getGithubUserInfoByAccessToken]get github user info err", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		seelog.Error("[getGithubUserInfoByAccessToken]read body err", err)
		return nil, err
	}
	var userInfo GithubUserInfo
	err = json.Unmarshal(body, &userInfo)
	return &userInfo, err
}

func ProfileGet(c *gin.Context) {
	sessionUser, exists := c.Get(ContextUserKey)
	if exists {
		HtmlSuccess(c, "admin/profile.html", gin.H{
			"user":     sessionUser,
			"comments": models.MustListUnreadComment(),
		})
	}
}

func ProfileUpdate(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	avatarUrl := c.PostForm("avatarUrl")
	nickName := c.PostForm("nickName")
	sessionUser, _ := c.Get(ContextUserKey)
	user, ok := sessionUser.(*models.User)
	if !ok {
		res["message"] = "server interval error"
		return
	}
	err = user.UpdateProfile(avatarUrl, nickName)
	if err != nil {
		seelog.Error("[ProfileUpdate]update profile err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
	res["user"] = models.User{AvatarUrl: avatarUrl, NickName: nickName}
}

func BindEmail(c *gin.Context) {
	var (
		err error
		res = gin.H{}
		subcriber forms.SubscribeForm
	)
	defer WriteJSON(c, res)
	if e := c.ShouldBind(&subcriber); e != nil {
		seelog.Error("[BindEmail]input param err", e)
		res["message"] = "input param err"
		return
	}
	email := subcriber.Email
	sessionUser, _ := c.Get(ContextUserKey)
	user, ok := sessionUser.(*models.User)
	if !ok {
		res["message"] = "server interval error"
		return
	}
	if len(user.Email) > 0 {
		res["message"] = "email have bound"
		return
	}
	_, err = models.GetUserByUsername(email)
	if err == nil {
		seelog.Error("[BindEmail]get user by username err", err)
		res["message"] = "email have be registered"
		return
	}
	err = user.UpdateEmail(email)
	if err != nil {
		seelog.Error("[BindEmail]update email err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func UnbindEmail(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	sessionUser, _ := c.Get(ContextUserKey)
	user, ok := sessionUser.(*models.User)
	if !ok {
		res["message"] = "server interval error"
		return
	}
	if user.Email == "" {
		res["message"] = "email haven't bound"
		return
	}
	err = user.UpdateEmail("")
	if err != nil {
		seelog.Error("[UnbindGithub]update github user info err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func UnbindGithub(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer WriteJSON(c, res)
	sessionUser, _ := c.Get(ContextUserKey)
	user, ok := sessionUser.(*models.User)
	if !ok {
		res["message"] = "server interval error"
		return
	}
	if user.GithubLoginId == "" {
		res["message"] = "github haven't bound"
		return
	}
	user.GithubLoginId = ""
	err = user.UpdateGithubUserInfo()
	if err != nil {
		seelog.Error("[UserLock]update github user info err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func UserIndex(c *gin.Context) {
	users, _ := models.ListUsers()
	user, _ := c.Get(ContextUserKey)
	HtmlSuccess(c, "admin/user.html", gin.H{
		"users":    users,
		"user":     user,
		"comments": models.MustListUnreadComment(),
	})
}

func UserLock(c *gin.Context) {
	var (
		err  error
		id  uint64
		res  = gin.H{}
		user *models.User
	)
	defer WriteJSON(c, res)
	id, err = ParseIdToUint(c.Param("id"), "UserLock")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	user, err = models.GetUser(uint(id))
	if err != nil {
		seelog.Error("[UserLock]get user err", err)
		res["message"] = err.Error()
		return
	}
	user.LockState = !user.LockState
	err = user.Lock()
	if err != nil {
		seelog.Error("[UserLock]lock user err", err)
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
