package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"ipmanlk/bettercopelk/components"
	"log"
	"net/http"

	"github.com/a-h/templ"
)

//go:embed assets
var assets embed.FS
var assetsFS, _ = fs.Sub(assets, "assets")

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsFS))))

	component := components.IndexPage()
	mux.Handle("GET /", templ.Handler(component))
	mux.Handle("POST /search", templ.Handler(components.ResultsList()))
	mux.Handle("GET /sse", http.HandlerFunc(sseHandler))
	fmt.Println("Listening on :3000")
	http.ListenAndServe(":3000", mux)
}

// http handler with SSE that sends 10 message with 5 seconds interval
func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()

	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Client disconnected")
			return
		default:
			// Create a bytes.Buffer to hold the rendered output
			var buf bytes.Buffer
			template := components.ResultItem(fmt.Sprintf("Title %d", i), fmt.Sprintf("Source %d", i))
			err := template.Render(r.Context(), &buf)
			if err != nil {
				log.Println(err)
				continue
			}
			fmt.Fprintf(w, "event: result\ndata: %s%s", buf.String(), "\n\n")
			w.(http.Flusher).Flush()
		}

	}

	fmt.Fprintf(w, "event: close\ndata: %s%s", "Closing connection", "\n\n")
}
