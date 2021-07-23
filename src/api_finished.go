// +build !active_file

package src

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (a *Api) RegisterRoutes(r *gin.Engine) {
	r.GET("/ping", a.Ping)
	r.GET("/leaderboard", a.Leaderboard)
}

func (a *Api) Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func (a *Api) Leaderboard(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an int"})
		return
	}

	out := []struct {
		Path  string `json:"path"`
		Count int    `json:"count"`
	}{}

	err = a.db.SelectContext(c.Request.Context(), &out, `
		SELECT path, COUNT(DISTINCT user_id) AS count
		FROM events
		GROUP BY 1
		ORDER BY 2 DESC
		LIMIT ?
	`, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, out)
}
