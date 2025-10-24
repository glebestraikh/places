package service

import (
	"context"
	"places/internal/model"
)

/*
	Интерфейсы бизнес логики для внешних потребителей
*/

// Service определяет интерфейс бизнес-логики
type Service interface {
	SearchLocations(ctx context.Context, query string) ([]model.Location, error)
	GetLocationDetails(ctx context.Context, location model.Location) (*model.LocationResult, error)
}

// GeocodingClient интерфейс для получения локаций
type GeocodingClient interface {
	GetLocations(ctx context.Context, query string) ([]model.Location, error)
}

// WeatherClient интерфейс для получения погоды
type WeatherClient interface {
	GetWeather(ctx context.Context, lat, lon float64) (*model.Weather, error)
}

// PlacesClient интерфейс для получения мест
type PlacesClient interface {
	GetPlaces(ctx context.Context, lat, lon, radius float64) ([]model.Place, error)
	GetPlaceDetails(ctx context.Context, xid string) (*model.Place, error)
}
