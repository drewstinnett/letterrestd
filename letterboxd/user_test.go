package letterboxd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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
