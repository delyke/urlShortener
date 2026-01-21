package audit

import "errors"

// Audit configuration validation errors.
var (
	// ErrAuditFileIsEmpty is returned when the audit file path is missing
	ErrAuditFileIsEmpty = errors.New("не указан путь к файлу аудита")
	// ErrAuditURLIsEmpty is returned when the audit URL is missing
	ErrAuditURLIsEmpty = errors.New("не указан URL внешней системы для аудита")
)
