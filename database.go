package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func createMysqlConnection() {
	cfg := mysql.NewConfig()

	cfg.Net = "tcp"
	cfg.User = getEnvMust("DATABASE_USER")
	cfg.Passwd = os.Getenv("DATABASE_PASSWORD")
	cfg.DBName = getEnvMust("DATABASE_SCHEMA")

	cfg.Addr = getEnvMust("DATABASE_HOST") +
		":" + getEnvDefault("DATABASE_PORT", "3306")

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
}

func createChunk(chunkId string, chunkSizeMB int64) {
	stmt, err := db.Prepare("INSERT INTO chunks (chunk_id, chunk_size_mb) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()
	_, err = stmt.Exec(chunkId, chunkSizeMB)
	if err != nil {
		log.Fatal(err)
	}
}

func createCollection(collection Collection) {
	stmt, err := db.Prepare("INSERT INTO collections (collection_id, file_original_name, file_mime_type, chunk_ids) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		collection.CollectionId,
		collection.OriginalName,
		collection.MimeType,
		strings.Join(collection.ChunkIds, ","),
	)

	if err != nil {
		log.Fatal(err)
	}
}

func getCollection(collectionId string) (collection *Collection) {
	stmt, err := db.Prepare("SELECT * FROM collections WHERE collection_id = ?")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	err = stmt.QueryRow(collectionId).Scan(collection)
	if err == sql.ErrNoRows {
		return nil
	}

	if err != nil {
		log.Fatal(err)
	}

	return
}

func purgeChunk(chunkLimit int64) (chunkIds []string) {
	chunkIds = []string{}

	stmt, err := db.Prepare("CALL purge_chunks_to_limit(?)")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()
	rows, err := stmt.Query(chunkLimit)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var chunkId string
		err = rows.Scan(&chunkId)

		if err != nil {
			log.Fatal(err)
		}

		chunkIds = append(chunkIds, chunkId)
	}

	return
}
