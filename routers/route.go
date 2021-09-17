package routers

import (
	"github.com/gin-gonic/gin"
	"blog/helpers"
	"path/filepath"
	"blog/system"
	"blog/models"
	"github.com/cihub/seelog"
	"net/http"
	"html/template"
	"blog/controllers"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"strings"
)

func ListTag() (tagStr string) {
	tags, err := models.ListTag()
	if err != nil {
		return
	}
	tagNames := make([]string, 0)
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	tagStr = strings.Join(tagNames, ",")
	return
}

func setTemplate(engine *gin.Engine) {

	funcMap := template.FuncMap{
		"dateFormat": helpers.DateFormat,
		"substring":  helpers.Substring,
		"isOdd":      helpers.IsOdd,
		"isEven":     helpers.IsEven,
		"truncate":   helpers.Truncate,
		"add":        helpers.Add,
		"minus":      helpers.Minus,
		"listTag":    ListTag,
	}

	engine.SetFuncMap(funcMap)
	engine.LoadHTMLGlob(filepath.Join(helpers.GetCurrentDirectory(), "/views/**/*"))
}

//setSessions initializes sessions & csrf middlewares
func setSessions(router *gin.Engine) {
	config := system.GetConfiguration()
	//https://github.com/gin-gonic/contrib/tree/master/sessions
	store := cookie.NewStore([]byte(config.SessionSecret))
	store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400, Path: "/"}) //Also set Secure: true if using SSL, you should though
	router.Use(sessions.Sessions("gin-session", store))
	//https://github.com/utrack/gin-csrf
	//router.Use(csrf.Middleware(csrf.Options{
	//	Secret: config.SessionSecret,
	//	ErrorFunc: func(c *gin.Context) {
	//		c.String(400, "CSRF token mismatch")
	//		c.Abort()
	//	},
	//}))
}

//+++++++++++++ middlewares +++++++++++++++++++++++

//SharedData fills in common data, such as user info, etc...
func SharedData() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if uID := session.Get(helpers.SessionKey); uID != nil {
			user, err := models.GetUser(uID)
			if err == nil {
				c.Set(helpers.ContextUserKey, user)
			}
		}
		if system.GetConfiguration().SignupEnabled {
			c.Set("SignupEnabled", true)
		}
		c.Next()
	}
}

//AuthRequired grants access to authenticated users, requires SharedData middleware
func AdminScopeRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, _ := c.Get(helpers.ContextUserKey); user != nil {
			if u, ok := user.(*models.User); ok && u.IsAdmin {
				c.Next()
				return
			}
		}
		seelog.Warnf("User not authorized to visit %s", c.Request.RequestURI)
		helpers.HandleMessage(c, http.StatusForbidden, "Forbidden!")
		c.Abort()
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, _ := c.Get(helpers.ContextUserKey); user != nil {
			if _, ok := user.(*models.User); ok {
				c.Next()
				return
			}
		}
		seelog.Warnf("User not authorized to visit %s", c.Request.RequestURI)
		helpers.HandleMessage(c, http.StatusForbidden,"Forbidden!")
		c.Abort()
	}
}

func InitRouter() *gin.Engine {
	router := gin.Default()

	setTemplate(router)
	setSessions(router)
	router.Use(SharedData())
	router.Static("/static", filepath.Join(helpers.GetCurrentDirectory(), "/static"))

	router.NoRoute(helpers.Handle404)
	router.GET("/", controllers.IndexGet)
	router.GET("/index", controllers.IndexGet)
	router.GET("/rss", controllers.RssGet)

	// user
	user := router.Group("/user")
	if system.GetConfiguration().SignupEnabled {
		user.GET("/register", controllers.RegisterGet)
		user.POST("/register", controllers.RegisterPost)
	}
	user.GET("/login", controllers.LoginGet)
	user.POST("/login", controllers.LoginPost)
	user.GET("/logout", controllers.LogoutGet)

	// third party login
	router.GET("/oauth2callback", controllers.Oauth2Callback)
	router.GET("/auth/:authType", controllers.AuthGet)

	// captcha
	router.GET("/captcha", controllers.CaptchaGet)

	visitor := router.Group("/visitor")
	visitor.Use(AuthRequired())
	{
		visitor.POST("/new_comment", controllers.CommentPost)
		visitor.POST("/comment/:id/delete", controllers.CommentDelete)
	}

	// subscriber
	router.GET("/subscribe", controllers.SubscribeGet)
	router.POST("/subscribe", controllers.Subscribe)
	router.GET("/active", controllers.ActiveSubscriber)
	router.GET("/unsubscribe", controllers.UnSubscribe)

	router.GET("/page/:id", controllers.PageGet)
	router.GET("/post/:id", controllers.PostGet)
	router.GET("/tag/:tag", controllers.TagGet)
	router.GET("/archives/:year/:month", controllers.ArchiveGet)

	router.GET("/link/:id", controllers.LinkGet)

	authorized := router.Group("/admin")
	authorized.Use(AdminScopeRequired())
	{
		// index
		authorized.GET("/index", controllers.AdminIndex)

		// image upload
		authorized.POST("/upload", controllers.Upload)

		// page
		authorized.GET("/page", controllers.PageIndex)
		authorized.GET("/new_page", controllers.PageNew)
		authorized.POST("/new_page", controllers.PageCreate)
		authorized.GET("/page/:id/edit", controllers.PageEdit)
		authorized.POST("/page/:id/edit", controllers.PageUpdate)
		authorized.POST("/page/:id/publish", controllers.PagePublish)
		authorized.POST("/page/:id/delete", controllers.PageDelete)

		// post
		authorized.GET("/post", controllers.PostIndex)
		authorized.GET("/new_post", controllers.PostNew)
		authorized.POST("/new_post", controllers.PostCreate)
		authorized.GET("/post/:id/edit", controllers.PostEdit)
		authorized.POST("/post/:id/edit", controllers.PostUpdate)
		authorized.POST("/post/:id/publish", controllers.PostPublish)
		authorized.POST("/post/:id/delete", controllers.PostDelete)

		// tag
		authorized.GET("/tag", controllers.TagIndex)
		authorized.POST("/new_tag", controllers.TagCreate)
		authorized.POST("/tag/:id/edit", controllers.TagUpdate)
		authorized.POST("/tag/:id/delete", controllers.TagDelete)

		// user
		authorized.GET("/user", controllers.UserIndex)
		authorized.POST("/user/:id/lock", controllers.UserLock)

		// profile
		authorized.GET("/profile", controllers.ProfileGet)
		authorized.POST("/profile", controllers.ProfileUpdate)
		authorized.POST("/profile/email/bind", controllers.BindEmail)
		authorized.POST("/profile/email/unbind", controllers.UnbindEmail)
		authorized.POST("/profile/github/unbind", controllers.UnbindGithub)

		// subscriber
		authorized.GET("/subscriber", controllers.SubscriberIndex)
		authorized.POST("/subscriber", controllers.SubscriberPost)

		// link
		authorized.GET("/link", controllers.LinkIndex)
		authorized.POST("/new_link", controllers.LinkCreate)
		authorized.POST("/link/:id/edit", controllers.LinkUpdate)
		authorized.POST("/link/:id/delete", controllers.LinkDelete)

		// comment
		authorized.POST("/comment/:id", controllers.CommentRead)
		authorized.POST("/read_all", controllers.CommentReadAll)

		// backup
		authorized.POST("/backup", controllers.BackupPost)
		authorized.POST("/restore", controllers.RestorePost)

		// mail
		authorized.POST("/new_mail", controllers.SendMail)
		authorized.POST("/new_batchmail", controllers.SendBatchMail)
	}
	return router
}