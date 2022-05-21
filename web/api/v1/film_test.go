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

func TestGetFilm(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/film/") {
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
	r.GET("/film/:id", v1.GetFilm)

	req, err := http.NewRequest(http.MethodGet, "/film/sweetback", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var ar v1.APIResponse
	content := w.Body.String()
	err = json.Unmarshal([]byte(content), &ar)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	film := ar.Data.(map[string]interface{})
	require.Equal(t, "Sweet Sweetback's Baadasssss Song", film["title"])

	// require.Equal(t, http.StatusOK, w.Code)
	// resp := &v1.APIResponse{}
	// err = json.Unmarshal(w.Body.Bytes(), &resp)
	// require.NoError(t, err)
	// f := resp.Data.(map[string]interface{})
	// require.NotNil(t, f)
	// require.Equal(t, "foo", f)
}
