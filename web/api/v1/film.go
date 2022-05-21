package v1

import (
	"github.com/apex/log"
	"github.com/drewstinnett/letterrestd/letterboxd"
	"github.com/gin-gonic/gin"
)

// ListExample godoc
// @Summary Get List Example
// @Schemes
// @Description Get a film from a film slug
// @Tags films
// @Accept json
// @Produce json
// @Param slug path string true "Film slug"
// @Success 200 {object} APIResponse
// @Router /films/{slug} [get]
func GetFilm(c *gin.Context) {
	slug := c.Param("slug")
	sc := c.MustGet("client").(letterboxd.ScrapeClient)
	film, err := sc.Film.Get(nil, slug)
	if err != nil {
		log.WithFields(log.Fields{
			"slug": slug,
		}).WithError(err).Warn("Error getting film")
		c.JSON(404, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.IndentedJSON(200, APIResponse{
		Data: film,
	})
}
