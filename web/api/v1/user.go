package v1

import (
	"github.com/drewstinnett/letterrestd/letterboxd"
	"github.com/gin-gonic/gin"
)

// UserExample godoc
// @Summary Get watched films per user
// @Schemes
// @Description Get watched fils of a user
// @Tags users
// @Accept json
// @Produce json
// @Param user path string true "user"
// @Success 200 {object} APIResponse
// @Router /users/{user}/watched [get]
func GetWatched(c *gin.Context) {
	user := c.Param("user")
	sc := c.MustGet("client").(letterboxd.ScrapeClient)
	films, _, err := sc.User.Watched(nil, user)
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
