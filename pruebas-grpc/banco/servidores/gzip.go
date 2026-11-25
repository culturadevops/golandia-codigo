package servidores

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type escritorGzip struct {
	http.ResponseWriter
	z *gzip.Writer
}

func (e escritorGzip) Write(p []byte) (int, error) { return e.z.Write(p) }

func gzipSiLaPiden(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			siguiente.ServeHTTP(w, r)
			return
		}
		z := gzip.NewWriter(w)
		defer z.Close()
		w.Header().Set("Content-Encoding", "gzip")
		siguiente.ServeHTTP(escritorGzip{w, z}, r)
	})
}
