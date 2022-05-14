package v1

import (
	"github.com/drewstinnett/letterrestd/letterboxd"
	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Data       interface{}            `json:"data"`
	Pagination *letterboxd.Pagination `json:"pagination,omitempty"`
}

func GetFilm(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
