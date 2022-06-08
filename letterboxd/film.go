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
	IMDB string `json:"imdb"`
	TMDB string `json:"tmdb"`
}

type Film struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Target      string           `json:"target"`
	ExternalIDs *ExternalFilmIDs `json:"external_ids,omitempty"`
}

type FilmService interface {
	GetExternalIDs(context.Context, *Film) error
	// GetFilmWithPath(*context.Context, string) (*Film, error)
	EnhanceFilmList(context.Context, *[]*Film) error
	Filmography(context.Context, *FilmographyOpt) ([]*Film, error)
	Get(context.Context, string) (*Film, error)
	ExtractFilmsWithPath(context.Context, string) ([]*Film, *Pagination, error)
	ExtractEnhancedFilmsWithPath(context.Context, string) ([]*Film, *Pagination, error)
	StreamBatch(context.Context, *FilmBatchOpts) (chan *Film, *Pagination, error)
	StreamBatchWithChan(context.Context, *FilmBatchOpts, chan *Film, chan error)
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
	profs := GetFilmographyProfessions()
	if !StringInSlice(f.Profession, profs) {
		return fmt.Errorf("Profession must be one of %v", profs)
	}
	return nil
}

type FilmBatchOpts struct {
	Watched   []string  `json:"watched"`
	Lists     []*ListID `json:"lists"`
	WatchList []string  `json:"watchlist"`
}

// StreamBatch Get a bunch of different films at once and stream them back to the user
func (f *FilmServiceOp) StreamBatchWithChan(ctx context.Context, batchOpts *FilmBatchOpts, filmsC chan *Film, done chan error) {
	// var films []*Film
	defer func() {
		log.Info("Completed Stream Batch")
		done <- nil
	}()
	var wg sync.WaitGroup

	// Handle User watched films first
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, username := range batchOpts.Watched {
			// userFilms := []Film{}
			log.WithFields(log.Fields{
				"username": username,
			}).Info("Fetching watched films")
			userFilmC := make(chan *Film)
			userDone := make(chan error)
			go f.client.User.StreamWatchedWithChan(ctx, username, userFilmC, userDone)
			loop := true
			for loop {
				select {
				case film := <-userFilmC:
					filmsC <- film
				case err := <-userDone:
					if err != nil {
						log.WithError(err).Error("Failed to get watched films")
						done <- err
					} else {
						log.Info("Finished")
						loop = false
					}
				}
			}
		}
	}()
	// Next up handle Lists
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, listID := range batchOpts.Lists {
			// userFilms := []Film{}
			log.WithFields(log.Fields{
				"username": listID.User,
				"slug":     listID.Slug,
			}).Info("Fetching list films")
			listFilmC := make(chan *Film)
			listDone := make(chan error)
			go f.client.User.StreamListWithChan(ctx, listID.User, listID.Slug, listFilmC, listDone)
			loop := true
			for loop {
				select {
				case film := <-listFilmC:
					filmsC <- film
				case err := <-listDone:
					if err != nil {
						log.WithError(err).Error("Failed to get list films")
						done <- err
					} else {
						log.Info("Finished")
						loop = false
					}
				}
			}
		}
	}()

	// Finally, handle watch lists
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, user := range batchOpts.WatchList {
			// userFilms := []Film{}
			log.WithFields(log.Fields{
				"username": user,
			}).Info("Fetching watchlist films")
			listFilmC := make(chan *Film)
			listDone := make(chan error)
			go f.client.User.StreamWatchListWithChan(ctx, user, listFilmC, listDone)
			loop := true
			for loop {
				select {
				case film := <-listFilmC:
					filmsC <- film
				case err := <-listDone:
					if err != nil {
						log.WithError(err).Error("Failed to get watchlist films")
						done <- err
					} else {
						log.Info("Finished")
						loop = false
					}
				}
			}
		}
	}()

	wg.Wait()
}

// StreamBatch Get a bunch of different films at once and stream them back to the user
func (f *FilmServiceOp) StreamBatch(ctx context.Context, batchOpts *FilmBatchOpts) (chan *Film, *Pagination, error) {
	retC := make(chan *Film, 1)

	var pagination *Pagination
	// Simple wg to make sure 1 thread has finished and we have the pagination data
	if len(batchOpts.Watched) > 0 {
		go func() {
			for _, username := range batchOpts.Watched {
				log.WithFields(log.Fields{
					"username": username,
				}).Info("Fetching watche films")
				var err error
				var watched chan []*Film
				watched, pagination, err = f.client.User.StreamWatched(ctx, username)
				if err != nil {
					log.WithFields(log.Fields{
						"username": username,
					}).Warn("Issue looking up watched films")
					return
				}
				log.Info("Got the channel, now cycle through em")
				for filmB := range watched {
					for _, film := range filmB {
						retC <- film
					}
				}

			}
		}()
	}
	return retC, pagination, nil
}

func (f *FilmServiceOp) ExtractFilmsWithPath(ctx context.Context, path string) ([]*Film, *Pagination, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s", path), nil)
	if err != nil {
		return nil, nil, err
	}
	firstItems, _, err := f.client.sendRequest(req, ExtractUserFilms)
	if err != nil {
		return nil, nil, err
	}
	films := firstItems.Data.([]*Film)
	return films, &firstItems.Pagintion, nil
}

func (f *FilmServiceOp) ExtractEnhancedFilmsWithPath(ctx context.Context, path string) ([]*Film, *Pagination, error) {
	films, pagination, err := f.ExtractFilmsWithPath(ctx, path)
	if err != nil {
		return nil, pagination, err
	}

	log.Debug("Launching EnhanceFilmList")
	err = f.client.Film.EnhanceFilmList(ctx, &films)
	if err != nil {
		return nil, nil, err
	}

	return films, pagination, nil
}

func (f *FilmServiceOp) Get(ctx context.Context, slug string) (*Film, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/film/%s", f.client.BaseURL, slug), nil)
	if err != nil {
		return nil, err
	}
	item, _, err := f.client.sendRequest(req, extractFilmFromFilmPage)
	if err != nil {
		return nil, err
	}
	return item.Data.(*Film), nil
}

func (f *FilmServiceOp) Filmography(ctx context.Context, opt *FilmographyOpt) ([]*Film, error) {
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

func (f *FilmServiceOp) EnhanceFilmList(ctx context.Context, films *[]*Film) error {
	var wg sync.WaitGroup
	wg.Add(len(*films))
	guard := make(chan struct{}, 5)
	for _, film := range *films {
		go func(film *Film) {
			defer wg.Done()
			guard <- struct{}{}
			log.Debugf("Looking up %v", film.Slug)
			if err := f.GetExternalIDs(ctx, film); err != nil {
				log.WithError(err).Warn("Failed to get external IDs")
				// return err
			}
			<-guard
		}(film)
	}
	wg.Wait()
	return nil
}

func (f *FilmServiceOp) GetExternalIDs(ctx context.Context, film *Film) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", f.client.BaseURL, film.Target), nil)
	if err != nil {
		return err
	}
	res, err := f.client.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
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
			ids.IMDB = extractIDFromURL(s.AttrOr("href", ""))
		}
		if val, ok := s.Attr("data-track-action"); ok && val == "TMDb" {
			ids.TMDB = extractIDFromURL(s.AttrOr("href", ""))
		}
	})

	return ids, nil
}

func extractFilmFromFilmPage(r io.Reader) (interface{}, *Pagination, error) {
	f := &Film{
		ExternalIDs: &ExternalFilmIDs{},
	}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, nil, err
	}
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if val, ok := s.Attr("property"); ok && val == "og:title" {
			fullTitle := s.AttrOr("content", "")
			f.Title = fullTitle[0 : len(fullTitle)-7]
		}
	})
	doc.Find("div").Each(func(i int, s *goquery.Selection) {
		s.Find("div").Each(func(i int, s *goquery.Selection) {
			if s.HasClass("poster film-poster") {
				if f.Slug == "" {
					f.Slug = normalizeSlug(s.AttrOr("data-film-slug", ""))
				}
				if f.Target == "" {
					f.Target = s.AttrOr("data-target-link", "")
				}
				if f.ID == "" {
					f.ID = s.AttrOr("data-film-id", "")
				}
			}
		})
	})
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if val, ok := s.Attr("data-track-action"); ok && val == "IMDb" {
			f.ExternalIDs.IMDB = extractIDFromURL(s.AttrOr("href", ""))
		}
		if val, ok := s.Attr("data-track-action"); ok && val == "TMDb" {
			f.ExternalIDs.TMDB = extractIDFromURL(s.AttrOr("href", ""))
		}
	})
	return f, nil, nil
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

func GetFilmographyProfessions() []string {
	return []string{"actor", "director", "producer", "writer"}
}
