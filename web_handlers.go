package main

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
)

//
// static endpoints

var indexViewHandler = httpLogFn(func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("./views", "index.html"))
})

var downloadViewHandler = httpLogFn(func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("./views", "download.html"))
})

var staticLibrariesHandler = httpLog(
	http.StripPrefix("/lib", http.FileServer(http.Dir("./node_modules"))),
)

var staticAssetsHandler = httpLog(
	http.StripPrefix("/assets", http.FileServer(http.Dir("./public"))),
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
	chunkOrderRaw := r.Header["X-Thepool-Chunk-Order"]
	if len(chunkOrderRaw) < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "no chunk order provided",
		})
		return
	}

	chunkOrder, err := strconv.ParseInt(chunkOrderRaw[0], 10, 64)

	if err != nil || chunkOrder < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid chunk order",
		})
		return
	}

	chunkSizeRaw := r.Header["X-Thepool-Chunk-Size"]
	if len(chunkOrderRaw) < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "no chunk size provided",
		})
		return
	}

	chunkSize, err := strconv.ParseInt(chunkSizeRaw[0], 10, 64)

	if err != nil || chunkSize < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid chunk size",
		})
		return
	}

	if chunkSize > WEBSERVER_SIZE_LIMIT_MB*MB {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Chunk size too large",
		})
		return
	}

	limitedReader := &io.LimitedReader{
		R: r.Body,
		N: WEBSERVER_SIZE_LIMIT_MB*MB + 1,
	}

	countableReader := &CountableReader{
		Reader: limitedReader,
	}

	chunkID := randID(10)
	err = uploadChunk(chunkID, chunkSize, countableReader)

	if err != nil {
		deleteChunks([]string{chunkID})
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if countableReader.Bytes != chunkSize {
		deleteChunks([]string{chunkID})
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{
			"success": false,
			"message": "payload size invalid",
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

	createChunk(chunkID, countableReader.Bytes, chunkOrder)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"id":      chunkID,
	})
})
