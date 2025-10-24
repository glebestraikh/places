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
	result := &model.LocationResult{
		Location: location,
	}

	// Создаем каналы для результатов
	weatherCh := make(chan *model.Weather, 1)
	placesCh := make(chan []model.Place, 1)
	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	// Асинхронно получаем погоду
	go func() {
		defer wg.Done()
		weather, err := s.weatherClient.GetWeather(ctx, location.Lat, location.Lon)
		if err != nil {
			errCh <- err
			return
		}
		weatherCh <- weather
	}()

	// Асинхронно получаем места
	go func() {
		defer wg.Done()
		places, err := s.placesClient.GetPlaces(ctx, location.Lat, location.Lon, 2000)
		if err != nil {
			errCh <- err
			return
		}

		// Асинхронно получаем детали для каждого места
		detailedPlaces := s.enrichPlacesWithDetails(ctx, places)
		placesCh <- detailedPlaces
	}()

	// Ждем завершения всех горутин
	wg.Wait()
	close(weatherCh)
	close(placesCh)
	close(errCh)

	// Собираем результаты
	if weather := <-weatherCh; weather != nil {
		result.Weather = weather
	}
	if places := <-placesCh; places != nil {
		result.Places = places
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
