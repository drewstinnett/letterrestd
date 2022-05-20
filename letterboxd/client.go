package letterboxd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/apex/log"
)

const (
	baseURL  = "https://letterboxd.com"
	maxPages = 50
)

type ScrapeClient struct {
	client    *http.Client
	UserAgent string
	// Config    ClientConfig
	BaseURL string
	User    UserService
	Film    FilmService
	List    ListService
	URL     URLService
	// Location  LocationService
	// Volume    VolumeService
}

type Response struct {
	*http.Response
}

// NewClient Generic new client creation
func NewScrapeClient(httpClient *http.Client) *ScrapeClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	userAgent := "letterrestd"
	c := &ScrapeClient{client: httpClient, UserAgent: userAgent, BaseURL: baseURL}

	// c.Location = &LocationServiceOp{client: c}
	// c.Volume = &VolumeServiceOp{client: c}
	c.User = &UserServiceOp{client: c}
	c.Film = &FilmServiceOp{client: c}
	c.List = &ListServiceOp{client: c}
	c.URL = &URLServiceOp{client: c}
	return c
}

type PageData struct {
	Data      interface{}
	Pagintion Pagination
}

func (c *ScrapeClient) sendRequest(req *http.Request, extractor func(io.Reader) (interface{}, *Pagination, error)) (*PageData, *Response, error) {
	res, err := c.client.Do(req)
	req.Close = true
	if err != nil {
		log.WithError(err).Warn("Error sending request")
		return nil, nil, err
	}
	// b, _ := ioutil.ReadAll(res.Body)

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes ErrorResponse
		b, _ := ioutil.ReadAll(res.Body)
		log.Debug(string(b))
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return nil, nil, errors.New(errRes.Message)
		}

		if res.StatusCode == http.StatusTooManyRequests {
			return nil, nil, fmt.Errorf("too many requests.  Check rate limit and make sure the userAgent is set right")
		} else if res.StatusCode == http.StatusNotFound {
			return nil, nil, fmt.Errorf("that entry was not found, are you sure it exists?")
		} else {
			return nil, nil, fmt.Errorf("error, status code: %d", res.StatusCode)
		}
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	items, pagination, err := extractor(bytes.NewReader(b))
	if err != nil {
		log.Warn("Error parsing response")
		return nil, nil, err
	}
	/*
		pagination, err := ExtractPaginationWithReader(bytes.NewReader(b))
		if err != nil {
			log.Warn("Error parsing pagination")
			return nil, nil, err
		}
	*/
	r := &Response{res}
	d := &PageData{
		Data: items,
	}
	if pagination != nil {
		d.Pagintion = *pagination
	}

	return d, r, nil
}

type ErrorResponse struct {
	Message string `json:"errors"`
}
