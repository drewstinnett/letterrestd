package letterboxd

import (
	"errors"
	"strings"
)

// 0 means undefined
// -1 means go for as far as you can!
func normalizeStartStop(firstPage, lastPage int) (int, int, error) {
	if firstPage == 0 && lastPage == 0 {
		return 1, 1, nil
	} else if firstPage == 0 {
		return 1, lastPage, nil
	} else if lastPage == 0 {
		return firstPage, firstPage, nil
	}
	if (lastPage >= 0) && (firstPage > lastPage) {
		return 0, 0, errors.New("last page must be greater than first page")
	}

	return firstPage, lastPage, nil
}

func normalizeSlug(slug string) string {
	slug = strings.TrimPrefix(slug, "/film/")
	slug = strings.TrimSuffix(slug, "/")
	return slug
}
