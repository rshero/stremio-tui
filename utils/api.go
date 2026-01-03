package apiutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	config "github.com/rshero/stremio-tui/config"
)

type ImdbSearchResult struct {
	Id            string `json:"id"`
	Type          string `json:"type"`
	PrimaryTitle  string `json:"primaryTitle"`
	OriginalTitle string `json:"originalTitle"`
}

type BehaviorHints struct {
	BingeGroup string `json:"bingeGroup"`
	VideoHash  string `json:"videoHash"`
	VideoSize  int64  `json:"videoSize"`
	Filename   string `json:"filename"`
}

type Season struct {
	Season       string `json:"season"`
	EpisodeCount int    `json:"episodeCount"`
}

type Episode struct {
	Id            string `json:"id"`
	Title         string `json:"title"`
	Season        string `json:"season"`
	EpisodeNumber int    `json:"episodeNumber"`
	Plot          string `json:"plot"`
	Rating        struct {
		AggregateRating float64 `json:"aggregateRating"`
		VoteCount       int     `json:"voteCount"`
	} `json:"rating"`
}

type AlcSearchResult struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Url           string        `json:"url"`
	BehaviorHints BehaviorHints `json:"behaviorHints"`
}

func ImdbSearch(query string, limit int) []ImdbSearchResult {
	apiUrl := fmt.Sprintf("%s/search/titles?query=%s&limit=%d", config.IMDB_API_URL, url.QueryEscape(query), limit)
	r, err := http.Get(apiUrl)
	if err != nil {
		fmt.Printf("Error making request: %v", err)
		return []ImdbSearchResult{}
	}
	defer r.Body.Close()

	var response struct {
		Titles []ImdbSearchResult `json:"titles"`
	}
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v", err)
		return []ImdbSearchResult{}
	}

	return response.Titles
}

func AlcStream(id string) []AlcSearchResult {
	apiUrl := fmt.Sprintf("%s%s.json", config.ALC_ADDON_URL, id)
	r, err := http.Get(apiUrl)
	if err != nil {
		fmt.Printf("Error making request: %v", err)
		return []AlcSearchResult{}
	}
	defer r.Body.Close()

	var response struct {
		Titles []AlcSearchResult `json:"streams"`
	}
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v", err)
		return []AlcSearchResult{}
	}

	return response.Titles
}

func FetchSeasons(id string) []Season {
	apiUrl := fmt.Sprintf("%s/titles/%s/seasons", config.IMDB_API_URL, id)
	r, err := http.Get(apiUrl)
	if err != nil {
		fmt.Printf("Error making request: %v", err)
		return []Season{}
	}
	defer r.Body.Close()

	var response struct {
		Seasons []Season `json:"seasons"`
	}
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v", err)
		return []Season{}
	}

	return response.Seasons
}

func FetchEpisodes(id string, season string) []Episode {
	apiUrl := fmt.Sprintf("%s/titles/%s/episodes?season=%s", config.IMDB_API_URL, id, season)
	r, err := http.Get(apiUrl)
	if err != nil {
		fmt.Printf("Error making request: %v", err)
		return []Episode{}
	}
	defer r.Body.Close()

	var response struct {
		Episodes []Episode `json:"episodes"`
	}
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v", err)
		return []Episode{}
	}

	return response.Episodes
}
