package main

type Chunk struct {
	ChunkId string `json:"id"`
	Size    int64  `json:"size"`
}

type Collection struct {
	CollectionId string
	OriginalName string
	MimeType     string
	ChunkIds     string
}

type CollectionPartial struct {
	OriginalName string `json:"originalName"`
	MimeType     string `json:"mimeType"`
	ChunkIds     string `json:"chunkIds"`
}

type CollectionDetails struct {
	CollectionId string  `json:"id"`
	OriginalName string  `json:"originalName"`
	MimeType     string  `json:"mimeType"`
	Chunks       []Chunk `json:"chunks"`
	TotalSize    int64   `json:"totalSize"`
}
