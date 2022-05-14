package v1

import (
	"github.com/drewstinnett/letterrestd/letterboxd"
	"github.com/gin-gonic/gin"
)

// @BasePath /api/v1

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description Get film information from the film slug
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /films/:id [get]
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
	c.IndentedJSON(200, APIResponse{
		Data: films,
	})
}
