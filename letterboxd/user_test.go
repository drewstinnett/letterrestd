package letterboxd

import (
	"os"
	"testing"

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
