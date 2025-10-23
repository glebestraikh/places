package openweather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

type openWeatherResponse struct {
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

func (c *Client) GetWeather(ctx context.Context, lat, lon float64) (*model.Weather, error) {
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric",
		lat, lon, c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status: %d", resp.StatusCode)
	}

	var owResp openWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&owResp); err != nil {
		return nil, err
	}

	weather := &model.Weather{
		Temp:      owResp.Main.Temp,
		FeelsLike: owResp.Main.FeelsLike,
		Humidity:  owResp.Main.Humidity,
		WindSpeed: owResp.Wind.Speed,
	}

	if len(owResp.Weather) > 0 {
		weather.Description = owResp.Weather[0].Description
		weather.Icon = owResp.Weather[0].Icon
	}

	return weather, nil
}
