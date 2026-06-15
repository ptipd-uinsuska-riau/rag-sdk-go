# RAG SDK Go

SDK Go untuk integrasi AI Chatbot RAG Universitas.

## Instalasi

```bash
go get github.com/ptipd-uinsuska-riau/rag-sdk-go
```

## Autentikasi

SDK mengirim header `X-App-Token` otomatis. Token **opsional**, tergantung konfigurasi RAG service:

1. **Publik** (default, `APP_TOKENS=[]`) — `appToken` boleh dikosongkan.
2. **Terbatas** — bila operator mengisi `APP_TOKENS`, isi `appToken` dengan token vendor dari admin.

## Penggunaan

### Chat Biasa

```go
package main

import (
    "context"
    "fmt"
    ragsdk "github.com/ptipd-uinsuska-riau/rag-sdk-go"
)

func main() {
    client := ragsdk.NewClient(
        "https://rag.universitas.ac.id/v1",
        "your-app-token",
    )

    resp, err := client.Chat(context.Background(), ragsdk.ChatRequest{
        Message: "Bagaimana cara mendaftar mahasiswa baru?",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("Jawaban:", resp.Answer)
    for _, src := range resp.Sources {
        fmt.Printf("Sumber: %s (%s)\n", src.Title, src.URL)
    }
}
```

### Streaming Chat

```go
err := client.ChatStream(ctx, ragsdk.ChatRequest{
    Message:   "Info beasiswa",
    SessionID: "session-123",
}, &ragsdk.SimpleStreamHandler{
    ChunkFn: func(text string) {
        fmt.Print(text)
    },
    SourcesFn: func(sources []ragsdk.ChatSource) {
        for _, s := range sources {
            fmt.Printf("\nSumber: %s\n", s.Title)
        }
    },
    DoneFn: func() {
        fmt.Println("\n--- Selesai ---")
    },
})
```

### Integrasi dengan Gin

```go
func chatHandler(c *gin.Context) {
    var req struct {
        Question string `json:"question"`
        UnitID   string `json:"unit_id"`
    }
    c.BindJSON(&req)

    client := ragsdk.NewClient(os.Getenv("RAG_URL"), os.Getenv("RAG_TOKEN"))

    resp, err := client.Chat(c.Request.Context(), ragsdk.ChatRequest{
        Message: req.Question,
        UnitID:  req.UnitID,
    })
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, resp)
}
```

### Integrasi dengan Fiber

```go
app.Post("/api/chat", func(c *fiber.Ctx) error {
    var body struct {
        Question string `json:"question"`
    }
    c.BodyParser(&body)

    client := ragsdk.NewClient(os.Getenv("RAG_URL"), os.Getenv("RAG_TOKEN"))

    resp, err := client.Chat(c.Context(), ragsdk.ChatRequest{
        Message: body.Question,
    })
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(resp)
})
```

## Opsi

```go
client := ragsdk.NewClient(url, token,
    ragsdk.WithLanguage("en"),
    ragsdk.WithTimeout(60 * time.Second),
    ragsdk.WithHTTPClient(customClient),
)
```

## Error Handling

```go
resp, err := client.Chat(ctx, req)
if errors.Is(err, ragsdk.ErrUnauthorized) {
    // Token tidak valid
} else if errors.Is(err, ragsdk.ErrRateLimited) {
    // Terlalu banyak request
} else if errors.Is(err, ragsdk.ErrServerError) {
    // Server error
}
```
