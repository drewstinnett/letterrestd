package letterboxd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/require"
)

func TestURLFilmographyBadProfession(t *testing.T) {
	client := NewScrapeClient(nil)
	_, err := client.URL.Items(nil, "https://www.letterboxd.com/televangelist/nicolas-cage")
	require.Error(t, err)
}

func TestURLFilmographyActor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/actor/nicolas-cage") {
			r, err := os.Open("testdata/filmography/actor/nicolas-cage.html")
			defer r.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, r)
			require.NoError(t, err)
			return
		} else if strings.HasPrefix(r.URL.Path, "/film/") {
			r, err := os.Open("testdata/film/sweetback.html")
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

	client := NewScrapeClient(nil)
	client.BaseURL = srv.URL

	items, err := client.URL.Items(nil, "https://www.letterboxd.com/actor/nicolas-cage")
	require.NoError(t, err)
	require.IsType(t, []*Film{}, items)
	require.Greater(t, len(items.([]*Film)), 0)
}

func TestNormalizeURLPath(t *testing.T) {
	tests := []struct {
		ourl         string
		expectedPath string
		wantErr      bool
		msg          string
	}{
		{"/film/everything-everywhere-all-at-once/", "/film/everything-everywhere-all-at-once", false, "no trailing slash"},
		{"/film/everything-everywhere-all-at-once", "/film/everything-everywhere-all-at-once", false, "trailing slash"},
		{"https://letterboxd.com/film/everything-everywhere-all-at-once/", "/film/everything-everywhere-all-at-once", false, "bare hostname"},
		{"https://www.letterboxd.com/film/everything-everywhere-all-at-once/", "/film/everything-everywhere-all-at-once", false, "www hostname"},
		{"https://www.google.com/film/everything-everywhere-all-at-once/", "", true, "invalid hostname"},
	}
	for _, tt := range tests {
		path, err := normalizeURLPath(tt.ourl)
		if tt.wantErr {
			require.Error(t, err, tt.msg)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.expectedPath, path, tt.msg)
		}
	}
}

func TestURLWatchList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/mondodrew/watchlist") {
			r, err := os.Open("testdata/user/watchlist.html")
			require.NoError(t, err)
			_, err = io.Copy(w, r)
			require.NoError(t, err)
			return
		} else if strings.HasPrefix(r.URL.Path, "/film/") {
			r, err := os.Open("testdata/film/sweetback.html")
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

	client := NewScrapeClient(nil)
	client.BaseURL = srv.URL

	items, err := client.URL.Items(nil, "https://www.letterboxd.com/mondodrew/watchlist")
	require.NoError(t, err)
	require.IsType(t, []*Film{}, items)
	require.Greater(t, len(items.([]*Film)), 0)
}

func TestURLUserList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "dave/list/official-top-250-narrative-feature-films") {
			r, err := os.Open("testdata/list/top250.html")
			require.NoError(t, err)
			_, err = io.Copy(w, r)
			require.NoError(t, err)
			return
		} else if strings.HasPrefix(r.URL.Path, "/film/") {
			r, err := os.Open("testdata/film/sweetback.html")
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

	client := NewScrapeClient(nil)
	client.BaseURL = srv.URL

	/*
		TODO: Need to mock this better
		items, err := client.URL.Items(nil, fmt.Sprintf("rboxd.com/dave/list/official-top-250-narrative-feature-films/")
		require.NoError(t, err)
		require.IsType(t, []*Film{}, items)
		require.Equal(t, len(items.([]*Film)), 250)
	*/
}
