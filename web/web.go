package web

import (
	"net/http"

	"github.com/drewstinnett/letterrestd/letterboxd"
	v1 "github.com/drewstinnett/letterrestd/web/api/v1"
	"github.com/gin-gonic/gin"
)

type RouterOpt struct {
	ScrapeClient *letterboxd.ScrapeClient
	Client       *http.Client
}

func NewRouter(r *RouterOpt) *gin.Engine {
	if r == nil {
		r = &RouterOpt{}
	}

	var hc *http.Client
	if r.Client == nil {
		hc = http.DefaultClient
	} else {
		hc = r.Client
	}

	var sc *letterboxd.ScrapeClient
	if r.ScrapeClient == nil {
		sc = letterboxd.NewScrapeClient(hc)
	} else {
		sc = r.ScrapeClient
	}

	router := gin.Default()
	router.Use(APIClient(sc))
	router.GET("/films/:id", v1.GetFilm)
	router.GET("/lists/:user/:slug", v1.GetList)
	return router
}

func APIClient(client *letterboxd.ScrapeClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("client", *client)
	}
}
