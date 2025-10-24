package geoapify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"places/internal/model"
	"strings"
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

// Geoapify Places API response
type geoapifyPlacesResponse struct {
	Features []struct {
		Properties struct {
			PlaceID    string   `json:"place_id"`
			Name       string   `json:"name"`
			Categories []string `json:"categories"`
			Datasource struct {
				Sourcename string `json:"sourcename"`
			} `json:"datasource"`
			Distance float64 `json:"distance"`
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

// Geoapify Place Details API response
type geoapifyDetailsResponse struct {
	Features []struct {
		Properties struct {
			PlaceID      string   `json:"place_id"`
			Name         string   `json:"name"`
			Categories   []string `json:"categories"`
			Street       string   `json:"street"`
			Housenumber  string   `json:"housenumber"`
			Postcode     string   `json:"postcode"`
			City         string   `json:"city"`
			State        string   `json:"state"`
			Country      string   `json:"country"`
			Formatted    string   `json:"formatted"`
			AddressLine1 string   `json:"address_line1"`
			AddressLine2 string   `json:"address_line2"`
			Website      string   `json:"website"`
			Datasource   struct {
				Sourcename string `json:"sourcename"`
				Raw        struct {
					Name         string `json:"name"`
					Description  string `json:"description"`
					Wikipedia    string `json:"wikipedia"`
					Wikidata     string `json:"wikidata"`
					Website      string `json:"website"`
					Phone        string `json:"phone"`
					OpeningHours string `json:"opening_hours"`
					Cuisine      string `json:"cuisine"`
					Image        string `json:"image"`
				} `json:"raw"`
			} `json:"datasource"`
			Contact struct {
				Phone string `json:"phone"`
				Email string `json:"email"`
			} `json:"contact"`
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

func (c *Client) GetPlaces(ctx context.Context, lat, lon, radius float64) ([]model.Place, error) {
	baseURL := "https://api.geoapify.com/v2/places"

	params := url.Values{}
	params.Add("categories", "tourism.sights,entertainment,catering,accommodation,commercial,leisure,sport")
	params.Add("filter", fmt.Sprintf("circle:%f,%f,%d", lon, lat, int(radius)))
	params.Add("bias", fmt.Sprintf("proximity:%f,%f", lon, lat))
	params.Add("limit", "50")
	params.Add("apiKey", c.apiKey)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geoapify places API returned status: %d", resp.StatusCode)
	}

	var gpResp geoapifyPlacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&gpResp); err != nil {
		return nil, err
	}

	places := make([]model.Place, 0, len(gpResp.Features))
	for _, feature := range gpResp.Features {
		props := feature.Properties
		if props.Name == "" {
			continue
		}

		coords := feature.Geometry.Coordinates
		placeLatitude, placeLongitude := 0.0, 0.0
		if len(coords) >= 2 {
			placeLongitude = coords[0]
			placeLatitude = coords[1]
		}

		categories := strings.Join(props.Categories, ", ")

		places = append(places, model.Place{
			Xid:      props.PlaceID,
			Name:     props.Name,
			Kinds:    categories,
			Lat:      placeLatitude,
			Lon:      placeLongitude,
			Distance: props.Distance,
		})
	}

	return places, nil
}

func (c *Client) GetPlaceDetails(ctx context.Context, placeID string) (*model.Place, error) {
	baseURL := "https://api.geoapify.com/v2/place-details"

	params := url.Values{}
	params.Add("id", placeID)
	params.Add("apiKey", c.apiKey)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geoapify place-details API returned status: %d", resp.StatusCode)
	}

	var pdResp geoapifyDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&pdResp); err != nil {
		return nil, err
	}

	if len(pdResp.Features) == 0 {
		return nil, fmt.Errorf("no details found for place id %s", placeID)
	}

	feature := pdResp.Features[0]
	props := feature.Properties

	categories := strings.Join(props.Categories, ", ")

	coords := feature.Geometry.Coordinates
	placeLatitude, placeLongitude := 0.0, 0.0
	if len(coords) >= 2 {
		placeLongitude = coords[0]
		placeLatitude = coords[1]
	}

	// Собираем описание из доступных полей
	description := props.Datasource.Raw.Description
	if description == "" && props.Datasource.Raw.Name != "" {
		description = props.Datasource.Raw.Name
	}

	// Добавляем адрес
	address := props.Formatted
	if address == "" {
		var addressParts []string
		if props.AddressLine1 != "" {
			addressParts = append(addressParts, props.AddressLine1)
		}
		if props.AddressLine2 != "" {
			addressParts = append(addressParts, props.AddressLine2)
		}
		if props.City != "" {
			addressParts = append(addressParts, props.City)
		}
		if props.State != "" {
			addressParts = append(addressParts, props.State)
		}
		if props.Postcode != "" {
			addressParts = append(addressParts, props.Postcode)
		}
		if props.Country != "" {
			addressParts = append(addressParts, props.Country)
		}
		if len(addressParts) > 0 {
			address = strings.Join(addressParts, ", ")
		}
	}

	if address != "" {
		if description != "" {
			description += "\n\n"
		}
		description += "Address: " + address
	}

	// Добавляем контактную информацию
	phone := props.Contact.Phone
	if phone == "" {
		phone = props.Datasource.Raw.Phone
	}
	if phone != "" {
		if description != "" {
			description += "\n"
		}
		description += fmt.Sprintf("Phone: %s", phone)
	}

	// Добавляем email
	if props.Contact.Email != "" {
		if description != "" {
			description += "\n"
		}
		description += fmt.Sprintf("Email: %s", props.Contact.Email)
	}

	// Добавляем часы работы
	if props.Datasource.Raw.OpeningHours != "" {
		if description != "" {
			description += "\n"
		}
		description += fmt.Sprintf("Opening hours: %s", props.Datasource.Raw.OpeningHours)
	}

	// Добавляем кухню (для ресторанов)
	if props.Datasource.Raw.Cuisine != "" {
		if description != "" {
			description += "\n"
		}
		description += fmt.Sprintf("Cuisine: %s", props.Datasource.Raw.Cuisine)
	}

	website := props.Datasource.Raw.Website
	if website == "" {
		website = props.Website
	}

	wikipedia := props.Datasource.Raw.Wikipedia

	// Получаем изображение из raw данных
	image := props.Datasource.Raw.Image

	place := &model.Place{
		Xid:         props.PlaceID,
		Name:        props.Name,
		Kinds:       categories,
		Lat:         placeLatitude,
		Lon:         placeLongitude,
		Description: description,
		Image:       image,
		Wikipedia:   wikipedia,
		WebSite:     website,
	}

	return place, nil
}
