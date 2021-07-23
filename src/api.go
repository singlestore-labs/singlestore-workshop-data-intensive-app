// +build active_file

package src

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) RegisterRoutes(r *gin.Engine) {
	r.GET("/ping", a.Ping)
}

func (a *Api) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
