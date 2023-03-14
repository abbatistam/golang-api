package models

// Errors struct
type Error struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}
