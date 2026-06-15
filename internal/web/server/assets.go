package server

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/thomiceli/opengist/public"
)

func embeddedAssetFS() (fs.FS, error) {
	return fs.Sub(public.Files, "assets")
}

func embeddedAssetHandler() (http.Handler, error) {
	assetsFS, err := embeddedAssetFS()
	if err != nil {
		return nil, err
	}
	return http.StripPrefix("/assets/", &mimeFileServer{http.FileServer(http.FS(assetsFS))}), nil
}

type mimeFileServer struct {
	http.Handler
}

func (m *mimeFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ext := path.Ext(r.URL.Path); ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
	}
	m.Handler.ServeHTTP(w, r)
}

func safeAssetName(name string) bool {
	return name != "" && !strings.Contains(name, "..")
}
