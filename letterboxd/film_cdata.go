package letterboxd

type CDATAActor struct {
	Type   string `json:"@type,omitempty"`
	Name   string `json:"name,omitempty"`
	SameAs string `json:"sameAs,omitempty"`
}
type CDATAAggregateRating struct {
	Type        string  `json:"@type,omitempty"`
	BestRating  float64 `json:"bestRating,omitempty"`
	Description string  `json:"description,omitempty"`
	RatingCount float64 `json:"ratingCount,omitempty"`
	RatingValue float64 `json:"ratingValue,omitempty"`
	ReviewCount float64 `json:"reviewCount,omitempty"`
	WorstRating float64 `json:"worstRating,omitempty"`
}
type CDATACountryOfOrigin struct {
	Type string `json:"@type,omitempty"`
	Name string `json:"name,omitempty"`
}
type CDATADirector struct {
	Type   string `json:"@type,omitempty"`
	Name   string `json:"name,omitempty"`
	SameAs string `json:"sameAs,omitempty"`
}
type CDATAProductionCompany struct {
	Type   string `json:"@type,omitempty"`
	Name   string `json:"name,omitempty"`
	SameAs string `json:"sameAs,omitempty"`
}

type CDATAReleasedEvent struct {
	Type      string `json:"@type,omitempty"`
	StartDate string `json:"startDate,omitempty"`
}

type CDATAFilm struct {
	ID                string                   `json:"@id,omitempty"`
	Type              string                   `json:"@type,omitempty"`
	Actors            []CDATAActor             `json:"actors,omitempty"`
	AggregateRating   CDATAAggregateRating     `json:"aggregateRating,omitempty"`
	CountryOfOrigin   []CDATACountryOfOrigin   `json:"countryOfOrigin,omitempty"`
	DateCreated       string                   `json:"dateCreated,omitempty"`
	DateModified      string                   `json:"dateModified,omitempty"`
	Director          []CDATADirector          `json:"director,omitempty"`
	Genre             []string                 `json:"genre,omitempty"`
	Image             string                   `json:"image,omitempty"`
	Name              string                   `json:"name,omitempty"`
	ProductionCompany []CDATAProductionCompany `json:"productionCompany,omitempty"`
	ReleasedEvent     []CDATAReleasedEvent     `json:"releasedEvent,omitempty"`
	URL               string                   `json:"url,omitempty"`
}
