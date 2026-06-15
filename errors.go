package ragsdk

import "errors"

var (
	// ErrUnauthorized — token tidak valid atau tidak aktif.
	ErrUnauthorized = errors.New("ragsdk: unauthorized — periksa X-App-Token Anda")

	// ErrRateLimited — terlalu banyak request, coba lagi nanti.
	ErrRateLimited = errors.New("ragsdk: rate limited — tunggu sebelum mengirim request lagi")

	// ErrServerError — kesalahan internal RAG service.
	ErrServerError = errors.New("ragsdk: server error — RAG service mengalami gangguan")
)
