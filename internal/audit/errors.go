package audit

import "errors"

var (
	// ErrAuditFileIsEmpty is returned when the audit file path is missing
	ErrAuditFileIsEmpty = errors.New("не указан путь к файлу аудита")
	// ErrAuditUrlIsEmpty is returned when the audit URL is missing
	ErrAuditUrlIsEmpty = errors.New("не указан URL внешней системы для аудита")
)
