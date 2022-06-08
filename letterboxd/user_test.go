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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestExtractUserFilms(t *testing.T) {
	f, err := os.Open("testdata/user/films.html")
	defer f.Close()
	require.NoError(t, err)

	items, _, err := ExtractUserFilms(f)
	films := items.([]*Film)
	// films = ret.([]FilmPreview)

	// log.Fatal(films)
	require.NoError(t, err)
	require.Greater(t, len(films), 70)
	require.Equal(t, "Cypress Hill: Insane in the Brain", films[0].Title)
}

func TestExtractUserFilmsSinglePage(t *testing.T) {
	f, err := os.Open("testdata/user/watched-films-single.html")
	defer f.Close()
	require.NoError(t, err)

	items, _, err := ExtractUserFilms(f)
	require.NoError(t, err)
	films := items.([]*Film)
	require.Equal(t, len(films), 34)
	require.Equal(t, "Irresistible", films[0].Title)
}

func TestExtractUser(t *testing.T) {
	f, err := os.Open("testdata/user/user.html")
	defer f.Close()
	require.NoError(t, err)
	user, _, err := ExtractUser(f)
	require.NoError(t, err)
	require.IsType(t, &User{}, user)
	u := user.(*User)
	require.Equal(t, "dankmccoy", u.Username)
	require.Equal(t, "Former writer for The Daily Show with Jon Stewart (also Trevor Noah). Podcaster -- The Flop House. I watch a lot of trash, but I also care about good stuff, I swear.", u.Bio)
}

func TestUserProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("testdata/user/user.html")
		defer f.Close()
		require.NoError(t, err)
		io.Copy(w, f)
	}))
	defer srv.Close()
	client := NewScrapeClient(nil)
	client.BaseURL = srv.URL

	item, _, err := client.User.Profile(nil, "dankmccoy")
	require.NoError(t, err)
	require.IsType(t, &User{}, item)
	require.Equal(t, 1398, item.WatchedFilmCount)
}

func TestUserProfileExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/dankmccoy") {
			f, err := os.Open("testdata/user/user.html")
			defer f.Close()
			require.NoError(t, err)
			io.Copy(w, f)
		}
	}))
	defer srv.Close()
	client := NewScrapeClient(nil)
	client.BaseURL = srv.URL

	tests := []struct {
		user   string
		expect bool
	}{
		{user: "dankmccoy", expect: true},
		{user: "neverexist", expect: false},
	}
	for _, tt := range tests {

		item, _, err := client.User.Profile(nil, tt.user)
		if tt.expect {
			require.NoError(t, err)
			require.IsType(t, &User{}, item)
		} else {
			require.Error(t, err)
		}
	}
}

func TestListWatched(t *testing.T) {
	sweetbackF, err := os.Open("testdata/film/sweetback.html")
	defer sweetbackF.Close()
	require.NoError(t, err)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/someguy/films/page/") {
			pageNo := strings.Split(r.URL.Path, "/")[4]
			rp, err := os.Open(fmt.Sprintf("testdata/user/watched-paginated/%v.html", pageNo))
			defer rp.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, rp)
			require.NoError(t, err)
			return
		} else if strings.HasPrefix(r.URL.Path, "/film/") {
			_, err = io.Copy(w, sweetbackF)
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

	watched, _, err := client.User.ListWatched(nil, "someguy")
	require.NoError(t, err)
	require.NotNil(t, watched)

	require.Equal(t, 321, len(watched))
}

func TestStreamWatchedWithChan(t *testing.T) {
	sweetbackF, err := os.Open("testdata/film/sweetback.html")
	defer sweetbackF.Close()
	require.NoError(t, err)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/someguy/films/page/") {
			pageNo := strings.Split(r.URL.Path, "/")[4]
			rp, err := os.Open(fmt.Sprintf("testdata/user/watched-paginated/%v.html", pageNo))
			defer rp.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, rp)
			require.NoError(t, err)
			return
		} else if strings.HasPrefix(r.URL.Path, "/film/") {
			_, err = io.Copy(w, sweetbackF)
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

	log.Info("Streaming movies")
	watchedC := make(chan *Film, 0)
	var watched []*Film
	done := make(chan error)
	go client.User.StreamWatchedWithChan(nil, "someguy", watchedC, done)
loop:
	for {
		select {

		case film := <-watchedC:
			watched = append(watched, film)
		case err := <-done:
			require.NoError(t, err)
			break loop
		default:
		}
	}

	require.NotEmpty(t, watched)
	require.Equal(t, 321, len(watched))
}
