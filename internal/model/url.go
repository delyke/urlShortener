package model

import "gorm.io/gorm"

type URL struct {
	gorm.Model
	OriginalURL string `json:"url"`
	ShortedURL  string `json:"shorted_url" gorm:"unique"`
}
