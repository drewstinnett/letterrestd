package letterboxd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractFilmExternalIDs(t *testing.T) {
	f, err := os.Open("testdata/film/sweetback.html")
	defer f.Close()
	require.NoError(t, err)

	ids, err := ExtractFilmExternalIDs(f)
	require.NoError(t, err)
	require.NotNil(t, ids)
	require.Equal(t, "tt0067810", ids.IMDBID)
	require.Equal(t, "5822", ids.TMDBID)
	// films := items.([]Film)
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
	film, err := extractFilmFromFilmPage(f)
	require.NoError(t, err)
	require.NotNil(t, film)
	require.NotNil(t, film.ExternalIDs)
	require.Equal(t, "tt0067810", film.ExternalIDs.IMDBID)
	require.Equal(t, "5822", film.ExternalIDs.TMDBID)
	require.Equal(t, "Sweet Sweetback's Baadasssss Song", film.Title)
}
