package main

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"
)

type AdminView struct {
	Users  []User
	Topics []string
}

type AdminData struct {
	Users     map[string][]Topic
	Exports   []string
	Completed float64
}

// handleAdmin serves the admin page.
func (s server) handleAdmin(c *gin.Context) {
	c.HTML(http.StatusOK, "admin.html", nil)
}

// handleAdminUsers serves the admin page to manage users.
func (s server) handleAdminUsers(c *gin.Context) {
	users, err := s.getUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	c.HTML(http.StatusOK, "users.html", AdminView{Users: users, Topics: s.topics})
}

// handleAdminAddUser handles adding a new user.
func (s server) handleAdminAddUser(c *gin.Context) {
	user := c.PostForm("user")
	topics := c.PostFormArray("topics[]")

	err := s.addUser(user, topics)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	s.handleAdminUsers(c)
}

// handleAdminRemoveUser handles removing a user.
func (s server) handleAdminRemoveUser(c *gin.Context) {
	user := c.PostForm("user")

	err := s.removeUser(user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	s.handleAdminUsers(c)
}

// handleAdminData handles the admin page summarising data.
func (s server) handleAdminData(c *gin.Context) {
	u, err := s.getUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	users := make(map[string][]Topic, len(u))
	for _, user := range u {
		topics, err := s.getTopics(user.Name)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", err)
			panic(err)
		}
		users[user.Name] = topics
	}

	assigned := 0.0
	completed := 0.0
	for _, v := range users {
		assigned += float64(len(v))
		for _, t := range v {
			if t.Completed {
				completed++
			}
		}
	}

	ratio := 0.0
	if assigned > 0 {
		ratio = (completed / assigned) * 100
	}

	// First, get a list of files in the directory.
	files, err := ioutil.ReadDir("./export")
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	exports := make([]string, len(files))
	for i, file := range files {
		exports[i] = file.Name()
	}

	c.HTML(http.StatusOK, "data.html", AdminData{
		Users:     users,
		Completed: ratio,
		Exports:   exports,
	})
}

// handleAdminDataExport handles downloading a completed report.
func (s server) handleAdminDataExport(c *gin.Context) {
	u, err := s.getUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	var topics []Topic
	for _, user := range u {
		t, err := s.getTopics(user.Name)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", err)
			panic(err)
		}
		topics = append(topics, t...)
	}

	filename := fmt.Sprintf("clef2019-qv-data-%d.csv", time.Now().Unix())
	f, err := os.OpenFile(path.Join("./export/", filename), os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	w := csv.NewWriter(f)
	for _, topic := range topics {
		err := w.Write([]string{topic.User, topic.Topic, topic.Query1, topic.Query2, topic.Filename})
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", err)
			panic(err)
		}
	}
	w.Flush()
	err = f.Close()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/export/%s", filename))
}

// handleAdminVoiceClips handles downloading the completed voice clips for topics by users.
func (s server) handleAdminVoiceClips(c *gin.Context) {
	u, err := s.getUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	w := zip.NewWriter(c.Writer)

	for _, user := range u {
		topics, err := s.getTopics(user.Name)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", err)
			panic(err)
		}
		for _, t := range topics {
			if !t.Completed {
				continue
			}

			fmt.Println(path.Join("./audio/", t.Filename))
			f, err := os.OpenFile(path.Join("./audio/", t.Filename), os.O_RDONLY, 0664)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "error.html", err)
				panic(err)
			}

			a, err := w.Create(t.Filename)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "error.html", err)
				panic(err)
			}

			b, err := ioutil.ReadAll(f)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "error.html", err)
				panic(err)
			}

			_, err = a.Write(b)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "error.html", err)
				panic(err)
			}

			err = f.Close()
			if err != nil {
				c.HTML(http.StatusInternalServerError, "error.html", err)
				panic(err)
			}
		}
	}

	err = w.Close()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	c.Writer.WriteHeader(200)
}
