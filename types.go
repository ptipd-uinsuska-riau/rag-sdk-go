package ragsdk

// ChatRequest berisi parameter untuk mengirim pesan ke RAG service.
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	UnitID    int    `json:"unit_id,omitempty"`
	Language  string `json:"language,omitempty"`
}

// ChatResponse berisi jawaban dari RAG service (response flat dari rag-be).
type ChatResponse struct {
	Answer           string         `json:"answer"`
	SessionID        string         `json:"session_id"`
	MessageID        string         `json:"message_id"`
	Sources          []ChatSource   `json:"sources"`
	Fallback         bool           `json:"fallback"`
	FallbackContacts map[string]any `json:"fallback_contacts"`
	Language         string         `json:"language"`
}

// ChatSource berisi referensi sumber jawaban.
type ChatSource struct {
	DocumentID   string  `json:"document_id,omitempty"`
	Title        string  `json:"title"`
	URL          string  `json:"url,omitempty"`
	ChunkPreview string  `json:"chunk_preview,omitempty"`
	RelevanceScore float64 `json:"relevance_score,omitempty"`
}

// StreamHandler menangani event SSE dari streaming chat.
type StreamHandler interface {
	OnChunk(text string)
	OnSources(sources []ChatSource)
	OnSessionId(sessionID string)
	// OnDone dipanggil saat stream selesai; messageID adalah id pesan asisten
	// di server (untuk feedback), kosong bila tidak tersedia.
	OnDone(messageID string)
	OnError(err error)
}

// Rating tipe feedback.
type Rating string

const (
	RatingHelpful    Rating = "helpful"
	RatingNotHelpful Rating = "not_helpful"
)

// ratingToInt memetakan Rating ke skala 1..5 yang diminta rag-be.
func ratingToInt(r Rating) int {
	if r == RatingHelpful {
		return 5
	}
	return 1
}

// FeedbackRequest berisi parameter feedback (rag-be: rating int 1..5).
type FeedbackRequest struct {
	Rating       int    `json:"rating"`
	FeedbackText string `json:"feedback_text,omitempty"`
}

// Option untuk konfigurasi client.
type Option func(*Client)

// streamEvent adalah payload JSON di tiap baris `data:` pada SSE rag-be.
type streamEvent struct {
	Type       string       `json:"type"`
	Content    string       `json:"content"`
	Message    string       `json:"message,omitempty"` // diisi pada event 'error'
	SessionID  string       `json:"session_id,omitempty"`
	MessageID  string       `json:"message_id,omitempty"`
	Sources    []ChatSource `json:"sources,omitempty"`
	IsFallback bool         `json:"is_fallback,omitempty"`
}
