package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//
// view endpoints

var indexViewHandler = httpLogFn(func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("./views", "index.html"))
})

var callbackViewHandler = httpLogFn(func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("./views", "callback.html"))
})

var downloadViewHandler = httpLogFn(func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	collection := getCollection(id)

	viewFile := filepath.Join("./views", "download.html")
	viewContentRaw, err := os.ReadFile(viewFile)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	viewContent := string(viewContentRaw)
	if collection == nil {
		collection = &CollectionDetails{
			OriginalName: "file not found",
			Chunks:       []Chunk{},
			MimeType:     "application/octet-stream",
			TotalSize:    0,
		}
	}

	viewContent = strings.ReplaceAll(viewContent, "{{collection_id}}", id)
	viewContent = strings.ReplaceAll(viewContent, "{{file_name}}", collection.OriginalName)
	viewContent = strings.ReplaceAll(viewContent, "{{chunk_count}}", fmt.Sprint(len(collection.Chunks)))
	viewContent = strings.ReplaceAll(viewContent, "{{total_size}}", byteSize(collection.TotalSize))
	viewContent = strings.ReplaceAll(viewContent, "{{mime_type}}", fmt.Sprint(collection.MimeType))

	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(viewContent))
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
	token := r.Header.Get("Authorization")
	isVaildToken := isVaildToken(token)

	writeJSON(w, http.StatusOK, map[string]any{
		"chunkSize":          WEBSERVER_SIZE_LIMIT_MB,
		"download":           WEBSERVER_STATIC_ENDPOINT_PREFIX,
		"isVaildToken":       isVaildToken,
		"authenticationLink": createAuthenticationLink(),
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
	token := r.Header.Get("Authorization")
	isVaildToken := isVaildToken(token)

	if !isVaildToken {
		writeJSON(w, http.StatusForbidden, map[string]any{
			"success": false,
			"message": "session token invalid",
		})
		return
	}

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
	token := r.Header.Get("Authorization")
	isVaildToken := isVaildToken(token)

	if !isVaildToken {
		writeJSON(w, http.StatusForbidden, map[string]any{
			"success": false,
			"message": "session token invalid",
		})
		return
	}

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

var getTokenHandler = httpLogFn(func(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	isUserInGuild := isUserInGuild(code)

	if !isUserInGuild {
		writeJSON(w, http.StatusForbidden, map[string]any{
			"success": false,
			"message": "user is not registered",
		})
		return
	}

	token := getVaildToken()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"token":   token,
	})
})
