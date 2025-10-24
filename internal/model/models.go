package model

// LocationResult представляет результат поиска локации с погодой и местами
type LocationResult struct {
	Location Location `json:"location"`
	Weather  *Weather `json:"weather"`
	Places   []Place  `json:"places"`
	Error    string   `json:"error,omitempty"`
}

// Location представляет географическую локацию
type Location struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"country,omitempty"`
	State   string  `json:"state,omitempty"`
}

// Weather представляет данные о погоде
type Weather struct {
	Temp        float64 `json:"temp"`
	FeelsLike   float64 `json:"feels_like"`
	Description string  `json:"description"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"wind_speed"`
	Icon        string  `json:"icon"`
}

// Place представляет интересное место
type Place struct {
	Xid         string  `json:"xid"`
	Name        string  `json:"name"`
	Kinds       string  `json:"kinds"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Distance    float64 `json:"distance,omitempty"`
	Description string  `json:"description,omitempty"`
	Image       string  `json:"image,omitempty"`
	Wikipedia   string  `json:"wikipedia,omitempty"`
}
