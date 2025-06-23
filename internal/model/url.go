package model

import "gorm.io/gorm"

type URL struct {
	gorm.Model
	OriginalUrl string `json:"url"`
	ShortedUrl  string `json:"shorted_url" gorm:"unique"`
}
