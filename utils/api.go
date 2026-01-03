package apiutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	config "github.com/rshero/stremio-tui/config"
)

const (
	maxRetries     = 5
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
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

func AlcStream(id string) ([]AlcSearchResult, error) {
	apiUrl := fmt.Sprintf("%s%s.json", config.ALC_ADDON_URL, id)

	backoff := initialBackoff
	for attempt := 0; attempt < maxRetries; attempt++ {
		r, err := http.Get(apiUrl)
		if err != nil {
			return []AlcSearchResult{}, fmt.Errorf("request failed: %v", err)
		}

		// Handle rate limiting (429)
		if r.StatusCode == http.StatusTooManyRequests {
			r.Body.Close()
			if attempt < maxRetries-1 {
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
			return []AlcSearchResult{}, fmt.Errorf("rate limited (429) after %d retries", maxRetries)
		}

		// Handle other HTTP errors
		if r.StatusCode != http.StatusOK {
			r.Body.Close()
			return []AlcSearchResult{}, fmt.Errorf("HTTP %d", r.StatusCode)
		}

		defer r.Body.Close()

		var response struct {
			Titles []AlcSearchResult `json:"streams"`
		}
		err = json.NewDecoder(r.Body).Decode(&response)
		if err != nil {
			return []AlcSearchResult{}, fmt.Errorf("decode error: %v", err)
		}

		return response.Titles, nil
	}

	return []AlcSearchResult{}, fmt.Errorf("max retries exceeded")
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
