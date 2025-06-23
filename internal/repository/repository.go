package repository

type URLRepository interface {
	Save(originalUrl string, shortedUrl string) error
	GetOriginalLink(shortedUrl string) (string, bool)
}
