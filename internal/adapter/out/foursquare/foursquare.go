package foursquare

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"places/internal/model"
	"strings"
)

type FoursquareClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewFoursquareClient(apiKey string) *FoursquareClient {
	return &FoursquareClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

type foursquareSearchResponse struct {
	Results []struct {
		Fsid     string `json:"fsq_id"`
		Name     string `json:"name"`
		Geocodes struct {
			Main struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"main"`
		} `json:"geocodes"`
		Categories []struct {
			Name string `json:"name"`
		} `json:"categories"`
		Distance int `json:"distance"`
	} `json:"results"`
}

type foursquareDetailsResponse struct {
	Fsid        string `json:"fsq_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Geocodes    struct {
		Main struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"main"`
	} `json:"geocodes"`
	Categories []struct {
		Name string `json:"name"`
	} `json:"categories"`
	Photos []struct {
		Prefix string `json:"prefix"`
		Suffix string `json:"suffix"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"photos"`
	Website string  `json:"website"`
	Tel     string  `json:"tel"`
	Rating  float64 `json:"rating"`
}

func (c *FoursquareClient) GetPlaces(ctx context.Context, lat, lon, radius float64) ([]model.Place, error) {
	url := fmt.Sprintf(
		"https://api.foursquare.com/v3/places/search?ll=%f,%f&radius=%d&limit=50",
		lat, lon, int(radius),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("foursquare API returned status: %d", resp.StatusCode)
	}

	var fsResp foursquareSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&fsResp); err != nil {
		return nil, err
	}

	places := make([]model.Place, 0, len(fsResp.Results))
	for _, p := range fsResp.Results {
		if p.Name != "" {
			categories := make([]string, 0, len(p.Categories))
			for _, cat := range p.Categories {
				categories = append(categories, cat.Name)
			}

			places = append(places, model.Place{
				Xid:      p.Fsid,
				Name:     p.Name,
				Kinds:    strings.Join(categories, ", "),
				Lat:      p.Geocodes.Main.Latitude,
				Lon:      p.Geocodes.Main.Longitude,
				Distance: float64(p.Distance),
			})
		}
	}

	return places, nil
}

func (c *FoursquareClient) GetPlaceDetails(ctx context.Context, fsid string) (*model.Place, error) {
	url := fmt.Sprintf(
		"https://api.foursquare.com/v3/places/%s?fields=fsq_id,name,description,geocodes,categories,photos,website,tel,rating",
		fsid,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("foursquare details API returned status: %d", resp.StatusCode)
	}

	var details foursquareDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	categories := make([]string, 0, len(details.Categories))
	for _, cat := range details.Categories {
		categories = append(categories, cat.Name)
	}

	image := ""
	if len(details.Photos) > 0 {
		photo := details.Photos[0]
		image = fmt.Sprintf("%soriginal%s", photo.Prefix, photo.Suffix)
	}

	description := details.Description
	if description == "" && details.Rating > 0 {
		description = fmt.Sprintf("Rating: %.1f/10", details.Rating)
	}
	if details.Tel != "" {
		if description != "" {
			description += "\n"
		}
		description += fmt.Sprintf("Tel: %s", details.Tel)
	}

	place := &model.Place{
		Xid:         details.Fsid,
		Name:        details.Name,
		Kinds:       strings.Join(categories, ", "),
		Lat:         details.Geocodes.Main.Latitude,
		Lon:         details.Geocodes.Main.Longitude,
		Description: description,
		Image:       image,
		Wikipedia:   details.Website,
	}

	return place, nil
}
