package graphhopper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"places/internal/model"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

type graphHopperResponse struct {
	Hits []struct {
		Point struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"point"`
		Name    string `json:"name"`
		Country string `json:"country"`
		State   string `json:"state"`
		City    string `json:"city"`
	} `json:"hits"`
}

func (c *Client) GetLocations(ctx context.Context, query string) ([]model.Location, error) {
	baseURL := "https://graphhopper.com/api/1/geocode"

	params := url.Values{}
	params.Add("q", query)
	params.Add("key", c.apiKey)
	params.Add("limit", "10")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding API returned status: %d", resp.StatusCode)
	}

	var ghResp graphHopperResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghResp); err != nil {
		return nil, err
	}

	locations := make([]model.Location, 0, len(ghResp.Hits))
	for _, hit := range ghResp.Hits {
		name := hit.Name
		if name == "" {
			name = hit.City
		}

		locations = append(locations, model.Location{
			Name:    name,
			Lat:     hit.Point.Lat,
			Lon:     hit.Point.Lng,
			Country: hit.Country,
			State:   hit.State,
		})
	}

	return locations, nil
}
