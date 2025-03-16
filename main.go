package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/team", func(c *gin.Context) {
		c.HTML(http.StatusOK, "team.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/projects", func(c *gin.Context) {
		c.HTML(http.StatusOK, "projects.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/join", func(c *gin.Context) {
		c.HTML(http.StatusOK, "join.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/donate", func(c *gin.Context) {
		c.HTML(http.StatusOK, "donate.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact.html", gin.H{
			"title": "Main website",
		})
	})
	r.GET("/footer", func(c *gin.Context) {
		c.HTML(http.StatusOK, "footer.html", gin.H{
			"title": "footer",
		})
	})
	r.Static("/static", "./static")
	r.Static("/templates", "./templates")

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in
	r.Run(":3000")
}
