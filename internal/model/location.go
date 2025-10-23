package model

// Location представляет географическую локацию
type Location struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"country,omitempty"`
	State   string  `json:"state,omitempty"`
}

// LocationResult представляет результат поиска локации с погодой и местами
type LocationResult struct {
	Location Location `json:"location"`
	Weather  *Weather `json:"weather"`
	Places   []Place  `json:"places"`
	Error    string   `json:"error,omitempty"`
}
