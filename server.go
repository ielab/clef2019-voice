package main

import (
	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type server struct {
	db       *bolt.DB
	topics   []string
	password string
}

type config struct {
	AdminPassword string `toml:"admin_password"`
}

func main() {
	var (
		s   server
		err error
	)

	var conf config
	_, err = toml.DecodeFile("config.toml", &conf)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll("./audio", 0774)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("./topics", 0774)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("./export", 0774)
	if err != nil {
		panic(err)
	}

	s.db, err = bolt.Open("conf.db", 0664, bolt.DefaultOptions)
	if err != nil {
		panic(err)
	}

	s.password = conf.AdminPassword
	s.topics, err = loadTopics("./topics")
	if err != nil {
		panic(err)
	}

	g := gin.Default()

	g.LoadHTMLGlob("./web/*.html")
	g.Static("/static", "./web/static")
	g.Static("/topics", "./topics")
	g.Static("/audio", "./audio")
	g.Static("/export", "./export")

	g.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", nil)
	})

	// A topic must first be listened to.
	// Consecutive handlers must receive post data which indicates progression through the topic.
	study := g.Group("/voice")
	{
		study.GET("/", s.studyAccessHandler)
		study.GET("/:user", s.studyHandler)
		study.GET("/:user/:topic/listen", s.listenHandler)      // User listens to topic, then speaks query.
		study.POST("/:user/:topic/write", s.writeHandler)       // User listens to their own spoken query, then writes it down.
		study.GET("/:user/:topic/write", s.writeHandler)        // User listens to their own spoken query, then writes it down.
		study.POST("/:user/:topic/confirm", s.confirmHandler)   // User listens to their own spoken query again, then writes it down again.
		study.POST("/:user/:topic/complete", s.completeHandler) // User completes topic and the two queries are checked to be the same.
	}
	// Middleware for preventing users without permission to a topic to access the topic page.
	study.Use(func(c *gin.Context) {
		_, err := s.getTopic(c.Query("user"), c.Param("topic"))
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", err)
			panic(err)
		}
		c.Next()
	})

	admin := g.Group("/admin", gin.BasicAuth(gin.Accounts{
		"admin": s.password,
	}))
	{
		admin.GET("/", s.handleAdmin)
		admin.GET("/user", s.handleAdminUsers)
		admin.POST("/user/add", s.handleAdminAddUser)
		admin.POST("/user/remove", s.handleAdminRemoveUser)
		admin.GET("/data", s.handleAdminData)
		admin.GET("/data/export", s.handleAdminDataExport)
		admin.GET("/data/audio.zip", s.handleAdminVoiceClips)
	}

	panic(http.ListenAndServe(":1313", g))
}

func loadTopics(dir string) ([]string, error) {
	// First, get a list of files in the directory.
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	topics := make([]string, len(files))
	for i, file := range files {
		topics[i] = strings.Replace(file.Name(), ".wav", "", -1)
	}

	return topics, nil
}
