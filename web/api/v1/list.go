package v1

import (
	"github.com/drewstinnett/letterrestd/letterboxd"
	"github.com/gin-gonic/gin"
)

// @BasePath /api/v1

// ListExample godoc
// @Summary Get List Example
// @Schemes
// @Description Get a list of films
// @Tags list
// @Accept json
// @Produce json
// @Param user path string true "Username of the list owner"
// @Param slug path string true "List slug"
// @Success 200 {object} APIResponse
// @Router /lists/{user}/{slug} [get]
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
