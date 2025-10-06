package main

type Collection struct {
	CollectionId string `json:"collectionId,omitempty"`
	OriginalName string `json:"originalName"`
	MimeType     string `json:"mimeType"`
	ChunkIds     string `json:"chunkIds"`
}
