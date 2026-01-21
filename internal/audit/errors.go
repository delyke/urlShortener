package audit

import "errors"

var (
	ErrAuditFileIsEmpty = errors.New("не указан путь к файлу аудита")
	ErrAuditURLIsEmpty  = errors.New("не указан URL внешней системы для аудита")
)
