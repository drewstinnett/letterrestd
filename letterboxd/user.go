package letterboxd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/apex/log"
)

type UserService interface {
	ListWatched(ctx *context.Context, userID string) ([]*Film, *Response, error)
	WatchList(ctx *context.Context, userID string) ([]*Film, *Response, error)
}

type UserServiceOp struct {
	client *ScrapeClient
}

func (u *UserServiceOp) WatchList(ctx *context.Context, userID string) ([]*Film, *Response, error) {
	var previews []*Film
	page := 1
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/watchlist/page/%d", u.client.BaseURL, userID, page), nil)
		if err != nil {
			return nil, nil, err
		}
		// var previews []FilmPreview
		items, resp, err := u.client.sendRequest(req, ExtractUserFilms)
		if err != nil {
			return nil, resp, err
		}
		partialFilms := items.Data.([]*Film)
		err = u.client.Film.EnhanceFilmList(ctx, &partialFilms)
		if err != nil {
			log.WithError(err).Warn("Failed to enhance film list")
		}
		previews = append(previews, partialFilms...)
		if items.Pagintion.IsLast {
			break
		}
		page++
	}
	return previews, nil, nil
}

func (u *UserServiceOp) ListWatched(ctx *context.Context, userID string) ([]*Film, *Response, error) {
	var previews []*Film
	page := 1
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/films/page/%d", u.client.BaseURL, userID, page), nil)
		if err != nil {
			return nil, nil, err
		}
		// var previews []FilmPreview
		items, resp, err := u.client.sendRequest(req, ExtractUserFilms)
		if err != nil {
			return nil, resp, err
		}
		partialFilms := items.Data.([]*Film)
		err = u.client.Film.EnhanceFilmList(ctx, &partialFilms)
		if err != nil {
			log.WithError(err).Warn("Failed to enhance film list")
		}
		previews = append(previews, partialFilms...)
		if items.Pagintion.IsLast {
			break
		}
		page++
	}
	return previews, nil, nil
}

func ExtractUserFilms(r io.Reader) (interface{}, *Pagination, error) {
	var previews []*Film
	var pageBuf bytes.Buffer
	tee := io.TeeReader(r, &pageBuf)
	doc, err := goquery.NewDocumentFromReader(tee)
	if err != nil {
		return nil, nil, err
	}
	doc.Find("li.poster-container").Each(func(i int, s *goquery.Selection) {
		s.Find("div").Each(func(i int, s *goquery.Selection) {
			if s.HasClass("film-poster") {
				f := Film{}
				f.ID = s.AttrOr("data-film-id", "")
				// f.Slug = s.AttrOr("data-film-slug", "")
				f.Slug = normalizeSlug(s.AttrOr("data-film-slug", ""))
				f.Target = s.AttrOr("data-target-link", "")
				// Real film name appears in the alt attribute for the poster
				s.Find("img.image").Each(func(i int, s *goquery.Selection) {
					f.Title = s.AttrOr("alt", "")
				})
				previews = append(previews, &f)
			}
		})
	})
	pagination, err := ExtractPaginationWithReader(&pageBuf)
	if err != nil {
		log.Warn("No pagination data found, assuming it to be a single page")
		pagination = &Pagination{
			CurrentPage: 1,
			NextPage:    1,
			TotalPages:  1,
			IsLast:      true,
		}
	}
	return previews, pagination, nil
}
