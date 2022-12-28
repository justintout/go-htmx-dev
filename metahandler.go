package gohtmxdev

import (
	"io"
	"mime"
	"net/http"

	_ "embed"
)

//go:embed hotreload.js
var hotreloadJS string

func Metahandler(reloader *Reloader) http.Handler {
	_ = mime.AddExtensionType(".js", "text/javascript")

	mux := http.NewServeMux()
	mux.HandleFunc("/_ghd/hotreload.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		io.WriteString(w, hotreloadJS)
	})
	mux.Handle("/_ghd/hotreload", reloader)
	return mux
}
