package webclient

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var embeddedDist embed.FS

func Handler() (http.Handler, error) {
	distFS, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return nil, err
	}

	fsHandler := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/client" || r.URL.Path == "/client/" {
			serveIndex(w, r, distFS)
			return
		}

		relPath := strings.TrimPrefix(r.URL.Path, "/client/")
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			relPath = "index.html"
		}

		if exists(distFS, relPath) {
			http.StripPrefix("/client/", fsHandler).ServeHTTP(w, r)
			return
		}

		serveIndex(w, r, distFS)
	}), nil
}

func serveIndex(w http.ResponseWriter, r *http.Request, fsys fs.FS) {
	indexHTML, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		http.Error(w, "open index.html", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(indexHTML)
}

func exists(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}
