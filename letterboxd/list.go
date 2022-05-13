package letterboxd

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type ListService interface {
	ListFilms(*context.Context, *ListFilmsOpt) ([]Film, error)
}

type ListServiceOp struct {
	client *ScrapeClient
}

// ListFilmsOpt is the options for the ListFilms method
type ListFilmsOpt struct {
	User      string // Username of the user for the list. Example: 'dave'
	Slug      string // Slug of the list: Example: 'official-top-250-narrative-feature-films'
	FirstPage int    // First page to fetch. Defaults to 1
	LastPage  int    // Last page to fetch. Defaults to FirstPage. Use -1 to fetch all pages
}

func (l *ListServiceOp) ListFilms(ctx *context.Context, opt *ListFilmsOpt) ([]Film, error) {
	var films []Film

	startPage, stopPage, err := normalizeStartStop(opt.FirstPage, opt.LastPage)
	if err != nil {
		return nil, err
	}

	page := startPage
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/list/%s/page/%d", l.client.BaseURL, opt.User, opt.Slug, page), nil)
		if err != nil {
			return nil, err
		}
		items, _, err := l.client.sendRequest(req, extractListFilms)
		if err != nil {
			return nil, err
		}
		films = append(films, items.Data.([]Film)...)
		if items.Pagintion.IsLast {
			break
		}
		// Set last page to the total number of pages if it's set to -1
		if opt.LastPage == -1 {
			stopPage = items.Pagintion.TotalPages
		}
		page++

		if (stopPage >= 0) && (page > stopPage) {
			break
		}

		if page >= maxPages {
			panic("Too many pages requested, close")
		}
	}
	return films, nil
}

func extractListFilms(r io.Reader) (interface{}, error) {
	var previews []Film
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	doc.Find("li.poster-container").Each(func(i int, s *goquery.Selection) {
		s.Find("div").Each(func(i int, s *goquery.Selection) {
			if s.HasClass("film-poster") {
				f := Film{}
				f.ID = s.AttrOr("data-film-id", "")
				f.Slug = normalizeSlug(s.AttrOr("data-film-slug", ""))
				f.Target = s.AttrOr("data-target-link", "")
				// Real film name appears in the alt attribute for the poster
				s.Find("img.image").Each(func(i int, s *goquery.Selection) {
					f.Title = s.AttrOr("alt", "")
				})
				previews = append(previews, f)
			}
		})
	})
	return previews, nil
}
