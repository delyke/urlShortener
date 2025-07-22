package model

type URL struct {
	UUID        string `json:"uuid"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}
