package server

// API response types — tygo generates TypeScript bindings from this file.

type Event struct {
	Name        string   `json:"name"`
	Sport       string   `json:"sport"`
	Location    string   `json:"location"`
	Date        string   `json:"date"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	ID          int      `json:"id"`
	PhotoCount  int      `json:"photo_count"`
	Galleries   int      `json:"galleries"`
}
