package main

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
)

//
// static endpoints

var indexViewHandler = httpLogFn(func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("./views", "index.html"))
})

var downloadViewHandler = httpLogFn(func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("./views", "download.html"))
})

var staticAssetsHandler = httpLog(
	http.StripPrefix("/assets/", http.FileServer(http.Dir("./node_modules"))),
)

//
// api endpoints

var configStatusHandler = httpLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"chunkSize": WEBSERVER_SIZE_LIMIT_MB,
		"download":  WEBSERVER_STATIC_ENDPOINT_PREFIX,
	})
}))

var getCollectionHandler = httpLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	collection := getCollection(id)

	if collection == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"success": false,
			"message": "collection not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    collection,
	})
}))

var createCollectionHandler = httpLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	id := randID(10)

	var payload Collection
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "payload invalid",
		})
		return
	}

	payload.CollectionId = id

	createCollection(payload)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"id":      id,
	})
}))

var createChunkHandler = httpLogFn(func(w http.ResponseWriter, r *http.Request) {
	limitedReader := &io.LimitedReader{
		R: r.Body,
		N: WEBSERVER_SIZE_LIMIT_MB*MB + 1,
	}

	countableReader := &CountableReader{
		Reader: limitedReader,
	}

	chunkID := randID(10)
	err := uploadChunk(chunkID, countableReader)

	if err != nil {
		deleteChunks([]string{chunkID})
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if countableReader.Bytes > WEBSERVER_SIZE_LIMIT_MB*MB {
		deleteChunks([]string{chunkID})
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{
			"success": false,
			"message": "payload too large",
		})
		return
	}

	createChunk(chunkID, countableReader.Bytes/MB+1)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"id":      chunkID,
	})
})
