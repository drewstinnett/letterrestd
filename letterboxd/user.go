package letterboxd

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type UserService interface {
	ListWatched(ctx *context.Context, userID string) ([]Film, *Response, error)
}

type UserServiceOp struct {
	client *Client
}

func (u *UserServiceOp) ListWatched(ctx *context.Context, userID string) ([]Film, *Response, error) {
	var previews []Film
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
		previews = append(previews, items.Data.([]Film)...)
		if items.Pagintion.IsLast {
			break
		}
		page++
	}
	return previews, nil, nil
}

func ExtractUserFilms(r io.Reader) (interface{}, error) {
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
				f.Slug = s.AttrOr("data-film-slug", "")
				f.Target = s.AttrOr("data-target-link", "")
				// Real film name appears in the alt attribute for the poster
				s.Find("img.image").Each(func(i int, s *goquery.Selection) {
					f.Title = s.AttrOr("alt", "")
				})
				previews = append(previews, f)
			}
		})
	})
	// v = &previews
	// log.Infof("%+v", v)
	// log.Infof("%+v", reflect.TypeOf(v))
	return previews, nil
}
