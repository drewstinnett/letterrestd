package letterboxd

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/require"
)

func TestExtractListFilms(t *testing.T) {
	f, err := os.Open("testdata/list/top250.html")
	defer f.Close()
	require.NoError(t, err)

	items, err := extractListFilms(f)
	films := items.([]Film)
	// films = ret.([]FilmPreview)

	// log.Fatal(films)
	require.NoError(t, err)
	require.Greater(t, len(films), 70)
	require.Equal(t, "Everything Everywhere All at Once", films[0].Title)
}

func TestListFilms(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/dave/list/official-top-250-narrative-feature-films/page/") {
			pageNo := strings.Split(r.URL.Path, "/")[5]
			r, err := os.Open(fmt.Sprintf("testdata/list/lists-page-%v.html", pageNo))
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
	client := NewScrapeClient(nil)
	client.BaseURL = srv.URL

	user := "dave"
	slug := "official-top-250-narrative-feature-films"

	tests := []struct {
		start     int
		stop      int
		wantCount int
	}{
		{1, 0, 100},
		{1, 2, 200},
		{1, 3, 250},
		{1, 4, 250},
		{1, -1, 250},
	}

	for _, tt := range tests {
		got, err := client.List.ListFilms(nil, &ListFilmsOpt{
			User:      user,
			Slug:      slug,
			FirstPage: tt.start,
			LastPage:  tt.stop,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, tt.wantCount, len(got))
	}
}
