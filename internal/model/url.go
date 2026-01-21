package model

// URL represents a shortened URL record.
type URL struct {
	UUID        string `json:"uuid"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	UserID      int64  `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}
