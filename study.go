package main

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s server) studyAccessHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func (s server) studyHandler(c *gin.Context) {
	user := c.Param("user")

	topics, err := s.getTopics(user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	c.HTML(http.StatusOK, "study.html", topics)
}

func (s server) listenHandler(c *gin.Context) {
	user := c.Param("user")
	topic := c.Param("topic")

	t, err := s.getTopic(user, topic)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	t.User = user

	c.HTML(http.StatusOK, "listen.html", t)
}

func (s server) writeHandler(c *gin.Context) {
	user := c.Param("user")
	topic := c.Param("topic")

	if c.Request.Method == "GET" {

		t, err := s.getTopic(user, topic)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", err)
			panic(err)
		}

		c.HTML(http.StatusOK, "write.html", t)
		return
	}

	// POST request.
	b, err := c.GetRawData()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	err = s.saveWAV(bytes.NewReader(b), topic, user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	c.Status(200)
}

func (s server) confirmHandler(c *gin.Context) {
	user := c.Param("user")
	topic := c.Param("topic")

	t, err := s.getTopic(user, topic)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	t.Query1 = c.PostForm("q")

	err = s.updateTopic(user, t)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	c.HTML(http.StatusOK, "confirm.html", t)
}

func (s server) completeHandler(c *gin.Context) {
	user := c.Param("user")
	topic := c.Param("topic")

	t, err := s.getTopic(user, topic)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	t.Query2 = c.PostForm("q")

	if t.Query1 != t.Query2 {
		t, err := s.getTopic(user, topic)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", err)
			panic(err)
		}

		t.Error = "The queries you have entered do not match. Please enter both of them again."

		c.HTML(http.StatusAccepted, "write.html", t)
		panic(err)
	}

	t.Completed = true

	err = s.updateTopic(user, t)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", err)
		panic(err)
	}

	c.HTML(http.StatusOK, "complete.html", t)
}
