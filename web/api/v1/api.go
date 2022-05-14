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

func GetList(c *gin.Context) {
	user := c.Param("user")
	slug := c.Param("slug")
	sc := c.MustGet("client").(letterboxd.ScrapeClient)
	films, err := sc.List.ListFilms(nil, &letterboxd.ListFilmsOpt{
		User:     user,
		Slug:     slug,
		LastPage: -1,
	})
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}
	//var data []*letterboxd.Film
	//for _, film := range films {
	//data = append(data, film)
	//}
	c.IndentedJSON(200, APIResponse{
		Data: films,
	})
}
