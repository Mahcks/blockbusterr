package omdb

import (
	"context"
	"fmt"

	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
)

type Service interface {
	GetMedia(ctx context.Context, imdbID string) (*Media, error)
}

type omdbService struct {
	gctx global.Context
	base *sling.Sling
}

func (o *omdbService) FetchAPIKeyFromDB(ctx context.Context) (string, error) {
	omdb, err := o.gctx.Crate().SQL.Queries().GetOMDbSettings(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get OMDb settings: %v", err)
	}

	if !omdb.APIKey.Valid || omdb.APIKey.String == "" {
		return "", fmt.Errorf("no OMDb API key found")
	}

	return omdb.APIKey.String, nil
}

type OMDbParams struct {
	APIKey string `url:"apikey"`
	IMDBID string `url:"i"`
}

type Media struct {
	Title      string   `json:"Title"`
	Year       string   `json:"Year"`
	Rated      string   `json:"Rated"`
	Released   string   `json:"Released"`
	Runtime    string   `json:"Runtime"`
	Genre      string   `json:"Genre"`
	Director   string   `json:"Director"`
	Writer     string   `json:"Writer"`
	Actors     string   `json:"Actors"`
	Plot       string   `json:"Plot"`
	Language   string   `json:"Language"`
	Country    string   `json:"Country"`
	Awards     string   `json:"Awards"`
	Poster     string   `json:"Poster"`
	Ratings    []Rating `json:"Ratings"`
	Metascore  string   `json:"Metascore"`
	IMDBRating string   `json:"imdbRating"`
	IMDBVotes  string   `json:"imdbVotes"`
	IMDBID     string   `json:"imdbID"`
	Type       string   `json:"Type"`
	DVD        string   `json:"DVD"`
	BoxOffice  string   `json:"BoxOffice"`
	Production string   `json:"Production"`
	Website    string   `json:"Website"`
	Response   string   `json:"Response"`
}

type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

func (o *omdbService) GetMedia(ctx context.Context, imdbID string) (*Media, error) {
	apiKey, err := o.FetchAPIKeyFromDB(ctx)
	if err != nil {
		return nil, err
	}

	var response Media
	_, err = o.base.New().Get("/").
		QueryStruct(OMDbParams{
			APIKey: apiKey,
			IMDBID: imdbID,
		}).
		ReceiveSuccess(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
