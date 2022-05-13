package letterboxd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ExternalFilmIDs struct {
	IMDBID string `json:"imdb_id"`
	TMDBID string `json:"tmdb_id"`
}

type Film struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Target      string           `json:"target"`
	ExternalIDs *ExternalFilmIDs `json:"external_ids,omitempty"`
}

type FilmService interface {
	GetExternalIDs(*context.Context, *Film) error
	GetFilmWithPath(*context.Context, string) (*Film, error)
}

type FilmServiceOp struct {
	client *ScrapeClient
}

func (f *FilmServiceOp) GetFilmWithPath(ctx *context.Context, path string) (*Film, error) {
	var film *Film
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", f.client.BaseURL, path), nil)
	if err != nil {
		return nil, err
	}
	_, err = f.client.client.Do(req)
	if err != nil {
		return nil, err
	}
	return film, nil
}

func (f *FilmServiceOp) GetExternalIDs(ctx *context.Context, film *Film) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", f.client.BaseURL, film.Target), nil)
	if err != nil {
		return err
	}
	res, err := f.client.client.Do(req)
	if err != nil {
		return err
	}
	ids, err := ExtractFilmExternalIDs(res.Body)
	if err != nil {
		return err
	}
	film.ExternalIDs = ids
	return nil
	// return f.client.sendRequest(req, ExtractFilmExternalIDs)
}

func ExtractFilmExternalIDs(r io.Reader) (*ExternalFilmIDs, error) {
	ids := &ExternalFilmIDs{}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if val, ok := s.Attr("data-track-action"); ok && val == "IMDb" {
			ids.IMDBID = extractIDFromURL(s.AttrOr("href", ""))
		}
		if val, ok := s.Attr("data-track-action"); ok && val == "TMDb" {
			ids.TMDBID = extractIDFromURL(s.AttrOr("href", ""))
			// ids.TMDBID = s.AttrOr("href", "")
		}
	})

	return ids, nil
}

func extractFilmFromFilmPage(r io.Reader) (*Film, error) {
	f := &Film{
		ExternalIDs: &ExternalFilmIDs{},
	}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if val, ok := s.Attr("property"); ok && val == "og:title" {
			fullTitle := s.AttrOr("content", "")
			f.Title = fullTitle[0 : len(fullTitle)-7]
		}
	})
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if val, ok := s.Attr("data-track-action"); ok && val == "IMDb" {
			f.ExternalIDs.IMDBID = extractIDFromURL(s.AttrOr("href", ""))
		}
		if val, ok := s.Attr("data-track-action"); ok && val == "TMDb" {
			f.ExternalIDs.TMDBID = extractIDFromURL(s.AttrOr("href", ""))
			// ids.TMDBID = s.AttrOr("href", "")
		}
	})
	return f, nil
}

func extractIDFromURL(url string) string {
	if strings.Contains(url, "imdb.com") {
		return strings.Split(url, "/")[4]
	} else if strings.Contains(url, "themoviedb.org") {
		return strings.Split(url, "/")[4]
	}
	return ""
}
