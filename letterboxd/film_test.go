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

func TestExtractFilmExternalIDs(t *testing.T) {
	f, err := os.Open("testdata/film/sweetback.html")
	require.NoError(t, err)
	defer f.Close()

	ids, err := ExtractFilmExternalIDs(f)
	require.NoError(t, err)
	require.NotNil(t, ids)
	require.Equal(t, "tt0067810", ids.IMDB)
	require.Equal(t, "5822", ids.TMDB)
}

func TestExtractFilmGenres(t *testing.T) {
	f, err := os.Open("testdata/film/sweetback.html")
	require.NoError(t, err)
	defer f.Close()

	got, err := ExtractFilmGenres(f)
	require.NoError(t, err)
	require.Equal(t, []string{"Crime", "Drama", "Action"}, got)
}

func TestExtractFilmThemes(t *testing.T) {
	f, err := os.Open("testdata/film/themes.html")
	require.NoError(t, err)
	defer f.Close()

	got, err := ExtractFilmThemes(f)
	require.NoError(t, err)
	require.Equal(t, []string{"Crude humor and satire", "Relationship comedy", "funny, comedy, humor, jokes or hilarious", "comedy, funny, hilarious, humor or jokes", "musical, songs, singing, comedy or funny", "comedy, horror, funny, humor or spooky", "action, villain, funny, humor or cool"}, got)
}

func TestExtractIDFromURL(t *testing.T) {
	tests := []struct {
		url string
		id  string
	}{
		{"http://www.imdb.com/title/tt0067810/maindetails", "tt0067810"},
		{"https://www.themoviedb.org/movie/5822/", "5822"},
		{"https://www.google.com", ""},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id := extractIDFromURL(tt.url)
			require.Equal(t, tt.id, id)
		})
	}
}

func TestExtractFilmFromFilmPage(t *testing.T) {
	f, err := os.Open("testdata/film/sweetback.html")
	defer f.Close()
	require.NoError(t, err)
	i, pagination, err := extractFilmFromFilmPage(f)
	film := i.(*Film)
	require.NoError(t, err)
	require.Nil(t, pagination)
	require.NotNil(t, film)
	require.NotNil(t, film.ExternalIDs)
	require.Equal(t, "tt0067810", film.ExternalIDs.IMDB)
	require.Equal(t, "5822", film.ExternalIDs.TMDB)
	require.Equal(t, "Sweet Sweetback's Baadasssss Song", film.Title)
	require.Equal(t, "sweet-sweetbacks-baadasssss-song", film.Slug)
	require.Equal(t, "/film/sweet-sweetbacks-baadasssss-song/", film.Target)
	require.Equal(t, "48640", film.ID)
}

func TestEnhanceFilmList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/dave/list/official-top-250-narrative-feature-films/page/") {
			pageNo := strings.Split(r.URL.Path, "/")[5]
			rp, err := os.Open(fmt.Sprintf("testdata/list/lists-page-%v.html", pageNo))
			defer rp.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, rp)
			require.NoError(t, err)
			return
		} else if strings.HasPrefix(r.URL.Path, "/film/") {
			rp, err := os.Open("testdata/film/sweetback.html")
			defer rp.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, rp)
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
	films, err := client.List.ListFilms(nil, &ListFilmsOpt{
		User:      user,
		Slug:      slug,
		FirstPage: 1,
		LastPage:  1,
	})
	require.NoError(t, err)
	require.NotNil(t, films)
	require.Equal(t, 100, len(films))

	// Make sure we don't get the external ids on a normal call
	// require.Nil(t, films[0].ExternalIDs)

	// Make sure we DO get them after enhancing
	// err = client.Film.EnhanceFilmList(nil, &films)
	// require.NoError(t, err)
	require.NotNil(t, films[0].ExternalIDs)
}

func TestFilmography(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/actor/nicolas-cage") {
			rp, err := os.Open("testdata/filmography/actor/nicolas-cage.html")
			defer rp.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, rp)
			require.NoError(t, err)
			return
		} else if strings.HasPrefix(r.URL.Path, "/film/") {
			rp, err := os.Open("testdata/film/sweetback.html")
			defer rp.Close()
			require.NoError(t, err)
			_, err = io.Copy(w, rp)
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

	profession := "actor"
	person := "nicolas-cage"
	films, err := client.Film.Filmography(nil, &FilmographyOpt{
		Person:     person,
		Profession: profession,
	})
	require.NoError(t, err)
	require.NotNil(t, films)
	require.Equal(t, 116, len(films))
	require.Equal(t, "Spider-Man: Into the Spider-Verse", films[0].Title)
}

func TestValidateFilmography(t *testing.T) {
	tests := []struct {
		opt     FilmographyOpt
		wantErr bool
	}{
		{FilmographyOpt{
			Profession: "actor",
		}, true},
		{FilmographyOpt{
			Person: "John Doe",
		}, true},
		{FilmographyOpt{
			Person:     "John Doe",
			Profession: "wait-staff",
		}, true},
		{FilmographyOpt{
			Person:     "nicolas-cage",
			Profession: "actor",
		}, false},
	}
	for _, tt := range tests {
		got := tt.opt.Validate()
		if tt.wantErr {
			require.Error(t, got)
		} else {
			require.NoError(t, got)
		}
	}
}

func TestStreamBatchWithChan(t *testing.T) {
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
		} else if strings.Contains(r.URL.Path, "/dave/list/official-top-250-narrative-feature-films/page/") {
			pageNo := strings.Split(r.URL.Path, "/")[5]
			rp, err := os.Open(fmt.Sprintf("testdata/list/lists-page-%v.html", pageNo))
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
	go client.Film.StreamBatchWithChan(nil, &FilmBatchOpts{
		Watched: []string{"someguy"},
		Lists: []*ListID{
			{"dave", "official-top-250-narrative-feature-films"},
		},
	}, watchedC, done)
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
	require.Equal(t, 571, len(watched))
}
