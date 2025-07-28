package repository

type URLRepository interface {
	Save(originalURL string, shortedURL string) error
	GetOriginalLink(shortedURL string) (string, error)
	Ping() error
}
