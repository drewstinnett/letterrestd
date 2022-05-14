package letterboxd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/apex/log"
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
	// GetFilmWithPath(*context.Context, string) (*Film, error)
	EnhanceFilmList(*context.Context, *[]*Film) error
	Filmography(*context.Context, *FilmographyOpt) ([]*Film, error)
}

type FilmServiceOp struct {
	client *ScrapeClient
}

type FilmographyOpt struct {
	Person     string // Person whos filmography is to be fetched
	Profession string // Profession of the person (actor, writer, director)
	// FirstPage  int    // First page to fetch. Defaults to 1
	// LastPage   int    // Last page to fetch. Defaults to FirstPage. Use -1 to fetch all pages
}

func (f *FilmographyOpt) Validate() error {
	if f.Person == "" {
		return fmt.Errorf("Person is required")
	}
	if f.Profession == "" {
		return fmt.Errorf("Profession is required")
	}
	profs := getFilmographyProfessions()
	if !StringInSlice(f.Profession, profs) {
		return fmt.Errorf("Profession must be one of %v", profs)
	}
	return nil
}

func (f *FilmServiceOp) Filmography(ctx *context.Context, opt *FilmographyOpt) ([]*Film, error) {
	var films []*Film
	err := opt.Validate()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/%s", f.client.BaseURL, opt.Profession, opt.Person), nil)
	if err != nil {
		return nil, err
	}
	items, _, err := f.client.sendRequest(req, extractFilmography)
	if err != nil {
		return nil, err
	}

	partialFilms := items.Data.([]*Film)

	// This is a bit costly, parallel time?
	err = f.client.Film.EnhanceFilmList(ctx, &partialFilms)
	if err != nil {
		log.WithError(err).Warn("Failed to enhance film list")
		return nil, err
	}

	films = append(films, partialFilms...)
	err = f.client.Film.EnhanceFilmList(ctx, &films)
	if err != nil {
		log.WithError(err).Warn("Failed to enhance film list")
		return nil, err
	}

	return films, nil
}

/*
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
*/

func (f *FilmServiceOp) EnhanceFilmList(ctx *context.Context, films *[]*Film) error {
	log.WithFields(log.Fields{
		"count": len(*films),
	}).Debug("Enhancing films")
	var wg sync.WaitGroup
	wg.Add(len(*films))
	for _, film := range *films {
		go func(film *Film) {
			defer wg.Done()

			if err := f.GetExternalIDs(ctx, film); err != nil {
				log.WithError(err).Warn("Failed to get external IDs")
				// return err
			}
		}(film)
	}
	wg.Wait()
	return nil
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

func extractFilmography(r io.Reader) (interface{}, *Pagination, error) {
	var previews []*Film
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, nil, err
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
				previews = append(previews, &f)
			}
		})
	})
	return previews, nil, nil
}

func getFilmographyProfessions() []string {
	return []string{"actor", "director", "producer", "writer"}
}
