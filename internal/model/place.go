package model

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
