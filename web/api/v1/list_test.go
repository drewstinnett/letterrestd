package v1_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/drewstinnett/letterrestd/letterboxd"
	"github.com/drewstinnett/letterrestd/web"
	v1 "github.com/drewstinnett/letterrestd/web/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestListFilms(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Warnf("%v", r.URL.Path)
		if strings.Contains(r.URL.Path, "/dave/list/official-top-250-narrative-feature-films/page/") {
			pageNo := strings.Split(r.URL.Path, "/")[5]
			r, err := os.Open(fmt.Sprintf("testdata/list/lists-page-%v.html", pageNo))
			defer r.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, r)
			require.NoError(t, err)
			return
		} else if strings.HasPrefix(r.URL.Path, "/film/") {
			r, err := os.Open(fmt.Sprintf("testdata/film/sweetback.html"))
			defer r.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, r)
			require.NoError(t, err)
			return
		} else {
			log.WithFields(log.Fields{
				"url": r.URL.String(),
			}).Warn("unexpected request")
			w.WriteHeader(http.StatusNotFound)
		}
		defer r.Body.Close()
	}))
	defer srv.Close()

	r := gin.Default()
	sc := letterboxd.NewScrapeClient(http.DefaultClient)
	sc.BaseURL = srv.URL
	r.Use(web.APIClient(sc))
	r.GET("/lists/:user/:slug", v1.GetList)

	req, err := http.NewRequest(http.MethodGet, "/lists/dave/official-top-250-narrative-feature-films", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	resp := &v1.APIResponse{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, 250, len(resp.Data.([]interface{})))
}

func TestListFilmsSinglePage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Warnf("%v", r.URL.Path)
		if strings.Contains(r.URL.Path, "/mondodrew/list/2022-movie-church") {
			r, err := os.Open("testdata/list/lists-single-page.html")
			defer r.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, r)
			require.NoError(t, err)
			return
		} else if strings.HasPrefix(r.URL.Path, "/film/") {
			r, err := os.Open(fmt.Sprintf("testdata/film/sweetback.html"))
			defer r.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, r)
			require.NoError(t, err)
			return
		} else {
			log.WithFields(log.Fields{
				"url": r.URL.String(),
			}).Warn("unexpected request")
			w.WriteHeader(http.StatusNotFound)
		}
		defer r.Body.Close()
	}))
	defer srv.Close()

	r := gin.Default()
	sc := letterboxd.NewScrapeClient(http.DefaultClient)
	sc.BaseURL = srv.URL
	r.Use(web.APIClient(sc))
	r.GET("/lists/:user/:slug", v1.GetList)

	req, err := http.NewRequest(http.MethodGet, "/lists/mondodrew/2022-movie-church", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	resp := &v1.APIResponse{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, 13, len(resp.Data.([]interface{})))
}
