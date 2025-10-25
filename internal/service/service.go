package service

import (
	"context"
	"places/internal/model"
	"sync"
)

type service struct {
	geocodingClient GeocodingClient
	weatherClient   WeatherClient
	placesClient    PlacesClient
}

// NewService создает новый экземпляр сервиса
func NewService(geocoding GeocodingClient, weather WeatherClient, places PlacesClient) Service {
	return &service{
		geocodingClient: geocoding,
		weatherClient:   weather,
		placesClient:    places,
	}
}

func (s *service) SearchLocations(ctx context.Context, query string) ([]model.Location, error) {
	return s.geocodingClient.GetLocations(ctx, query)
}

func (s *service) GetLocationDetails(ctx context.Context, location model.Location) (*model.LocationResult, error) {
	result := &model.LocationResult{Location: location}

	weatherCh := make(chan *model.Weather, 1)
	placesCh := make(chan []model.Place, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	// Погода
	go func() {
		defer wg.Done()
		if w, err := s.weatherClient.GetWeather(ctx, location.Lat, location.Lon); err == nil {
			weatherCh <- w
		}
	}()

	// Места
	go func() {
		defer wg.Done()
		if ps, err := s.placesClient.GetPlaces(ctx, location.Lat, location.Lon, 2000); err == nil {
			placesCh <- s.enrichPlacesWithDetails(ctx, ps)
		}
	}()

	// Закрываем каналы, когда все писатели завершились
	wg.Wait()
	close(weatherCh)
	close(placesCh)

	// Читаем без риска блокировки
	if w, ok := <-weatherCh; ok && w != nil {
		result.Weather = w
	}
	if ps, ok := <-placesCh; ok && ps != nil {
		result.Places = ps
	}

	return result, nil
}

func (s *service) enrichPlacesWithDetails(ctx context.Context, places []model.Place) []model.Place {
	if len(places) == 0 {
		return places
	}

	detailedPlaces := make([]model.Place, len(places))
	var wg sync.WaitGroup // ждёт завершения всех горутин
	var mu sync.Mutex     // мьютекс, чтобы безопасно записывать данные в общий массив detailedPlaces

	for i, place := range places {
		wg.Add(1)
		go func(idx int, p model.Place) {
			defer wg.Done()

			details, err := s.placesClient.GetPlaceDetails(ctx, p.Xid)
			if err == nil && details != nil {
				mu.Lock()
				detailedPlaces[idx] = *details
				mu.Unlock()
			} else {
				mu.Lock()
				detailedPlaces[idx] = p
				mu.Unlock()
			}
		}(i, place)
	}

	wg.Wait()
	return detailedPlaces
}
