# rag-sdk-go — SDK Go untuk Chatbot RAG UIN Suska Riau

Klien Go untuk layanan [rag-be](../rag-be) (`/v1/chat`, `/v1/chat/stream`,
feedback). Library — bukan service yang di-deploy.

- Module: `github.com/ptipd-uinsuska-riau/rag-sdk-go` · Go **1.22+** · tag `v0.1.0`

## Instalasi
```bash
go get github.com/ptipd-uinsuska-riau/rag-sdk-go@v0.1.0
```

## Konfigurasi
Tidak ada file config. Parameter diberikan ke `NewClient`:
- `ragBaseURL` — mis. `https://rag.uin-suska.ac.id/v1`
- `appToken` — nilai `X-App-Token`. Boleh kosong bila `APP_TOKENS` di rag-be kosong (publik); **wajib** bila rag-be memproteksi endpoint.
- Opsi: `WithLanguage("id"|"en")`, `WithTimeout(d)`, `WithHTTPClient(hc)`.

## Penggunaan
```go
import (
    "context"
    "fmt"
    ragsdk "github.com/ptipd-uinsuska-riau/rag-sdk-go"
)

client := ragsdk.NewClient("https://rag.uin-suska.ac.id/v1", "" /* appToken */)

// Non-streaming
resp, err := client.Chat(context.Background(), ragsdk.ChatRequest{
    Message: "Bagaimana cara mendaftar mahasiswa baru?",
})
if err != nil { panic(err) }
fmt.Println(resp.Answer)

// Streaming (SSE)
_ = client.ChatStream(context.Background(),
    ragsdk.ChatRequest{Message: "Apa itu PPID?"},
    &ragsdk.SimpleStreamHandler{
        OnChunk: func(s string) { fmt.Print(s) },
    },
)

// Feedback
_ = client.Feedback(context.Background(), resp.MessageID, ragsdk.RatingHelpful, "")
```

### API
- `NewClient(ragBaseURL, appToken string, opts ...Option) *Client`
- `(*Client) Chat(ctx, ChatRequest) (*ChatResponse, error)`
- `(*Client) ChatStream(ctx, ChatRequest, StreamHandler) error`
- `(*Client) Feedback(ctx, messageID string, rating Rating, comment string) error`
- Tipe: `ChatRequest{Message, SessionID, UnitID, Language}`, `ChatResponse{Answer, SessionID, MessageID, Sources, Fallback, ...}`, `ChatSource`, `StreamHandler`/`SimpleStreamHandler`, `Rating` (`RatingHelpful`/`RatingNotHelpful`).
- Error: `ErrUnauthorized`, `ErrRateLimited` (HTTP 429), `ErrServerError`.

> Untuk percakapan multi-turn, simpan `resp.SessionID` lalu kirim balik di `ChatRequest.SessionID`. Hormati `ErrRateLimited` (rate limit rag-be: chat 30/menit/IP).
