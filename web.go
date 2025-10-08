package main

import (
	"crypto/tls"
	"log"
	"net/http"
)

func openWebserver() {
	mux := http.NewServeMux()

	mux.Handle("GET /", indexViewHandler)
	mux.Handle("GET /{id}", downloadViewHandler)
	mux.Handle("GET /lib/{path...}", staticLibrariesHandler)
	mux.Handle("GET /assets/{path...}", staticAssetsHandler)

	mux.Handle("GET /api/config", configStatusHandler)
	mux.Handle("GET /api/collections/{id}", getCollectionHandler)
	mux.Handle("POST /api/collections", createCollectionHandler)
	mux.Handle("POST /api/chunks", createChunkHandler)

	tlsCfg := &tls.Config{
		NextProtos: []string{"h2", "http/1.1"},
		MinVersion: tls.VersionTLS12,
	}

	srv := &http.Server{
		Addr:         ":" + WEBSERVER_PORT,
		Handler:      mux,
		ReadTimeout:  0,
		WriteTimeout: 0,
		TLSConfig:    tlsCfg,
	}

	cert := "/tmp/cert.pem"
	key := "/tmp/key.pem"

	log.Printf("listening on https://localhost:%s", WEBSERVER_PORT)
	log.Fatal(srv.ListenAndServeTLS(cert, key))
}
