package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func openWebserver() {
	port := getEnvDefault("WEBSERVER_PORT", "8080")
	sizeLimitMBRaw := getEnvDefault("WEBSERVER_SIZE_LIMIT_MB", "128")

	sizeLimitMB, err := strconv.ParseInt(sizeLimitMBRaw, 10, 64)
	if err != nil || sizeLimitMB <= 0 {
		log.Fatalf("WEBSERVER_SIZE_LIMIT_MB=%q invalid", sizeLimitMBRaw)
	}
	mux := http.NewServeMux()

	publicDir := "./public"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path

		if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/assets/") {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(publicDir, "index.html"))
	})

	assets := http.StripPrefix("/assets/", http.FileServer(http.Dir("./node_modules")))
	mux.Handle("/assets/", logmw(assets))

	mux.Handle("/api/config", logmw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"chunkSize": sizeLimitMB})
	})))

	mux.Handle("/api/collections/", logmw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		id := filepath.Base(r.URL.Path)
		collection := getCollection(id)
		if collection == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"success": false, "message": "collection not found",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": collection})
	})))

	mux.Handle("/api/collections", logmw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var payload Collection
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Println(err.Error())
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"success": false, "message": "payload invalid",
			})
			return
		}
		payload.CollectionId = randID(10)
		createCollection(payload)
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "id": payload.CollectionId})
	})))

	mux.Handle("/api/chunks", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		const MB = int64(1024 * 1024)
		maxBytes := sizeLimitMB * MB

		lr := &io.LimitedReader{R: r.Body, N: maxBytes + 1}

		pr, pw := io.Pipe()
		var n int64

		go func() {
			defer pw.Close()
			buf := make([]byte, 256*1024)
			for {
				nr, er := lr.Read(buf)
				if nr > 0 {
					n += int64(nr)
					if _, ew := pw.Write(buf[:nr]); ew != nil {
						_ = pw.CloseWithError(ew)
						return
					}
				}
				if er != nil {
					if er != io.EOF {
						_ = pw.CloseWithError(er)
					}
					return
				}
			}
		}()

		chunkID := randID(10)

		if err := uploadChunk(chunkID, -1, pr); err != nil {
			deleteChunk([]string{chunkID})
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": err.Error()})
			return
		}

		if n > maxBytes {
			deleteChunk([]string{chunkID})
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{"success": false, "message": "chunk too large"})
			return
		}

		sizeMB := (n + MB - 1) / MB
		createChunk(chunkID, sizeMB)

		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"id":      chunkID,
		})
	}))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  0,
		WriteTimeout: 0,
		TLSConfig: &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
			MinVersion: tls.VersionTLS12,
		},
	}

	cert := "/tmp/cert.pem"
	key := "/tmp/key.pem"

	log.Printf("listening on https://localhost:%s", port)
	log.Fatal(srv.ListenAndServeTLS(cert, key))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func logmw(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
}
