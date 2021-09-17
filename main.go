package main

import (
	"flag"
	"github.com/cihub/seelog"
	"github.com/claudiu/gocron"
	"blog/controllers"
	"blog/models"
	"blog/system"
	"blog/routers"
	"github.com/gin-gonic/gin"
	"blog/helpers"
)

func main() {
	path := helpers.GetCurrentDirectory()
	configFilePath := flag.String("C", path + "/conf/conf.yaml", "config file path")
	logConfigPath := flag.String("L", path + "/conf/seelog.xml", "log config file path")
	flag.Parse()

	logger, err := seelog.LoggerFromConfigAsFile(*logConfigPath)
	if err != nil {
		seelog.Critical("[main]err parsing seelog config file", err)
		return
	}
	seelog.ReplaceLogger(logger)
	defer seelog.Flush()

	if err := system.LoadConfiguration(*configFilePath); err != nil {
		seelog.Critical("[main]err parsing config log file", err)
		return
	}

	db, err := models.InitDB()
	if err != nil {
		seelog.Critical("[main]err open databases", err)
		return
	}
	defer db.Close()

	// todo 生产环境要设置为ReleaseMode
	gin.SetMode(gin.DebugMode)

	//Periodic tasks
	gocron.Every(1).Day().Do(controllers.CreateXMLSitemap)
	gocron.Every(7).Days().Do(controllers.Backup)
	gocron.Start()

	router := routers.InitRouter()
	router.Run(system.GetConfiguration().Addr)
}
