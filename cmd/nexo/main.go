package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/yecharlot/NEXO/internal/httpapi"
	"github.com/yecharlot/NEXO/internal/store"
)

func main() {
	dataDir := env("NEXO_DATA_DIR", "./data")
	webDir := env("NEXO_WEB_DIR", "./web")
	addr := env("NEXO_ADDR", ":8090")
	if err := os.MkdirAll(filepath.Join(dataDir, "media"), 0700); err != nil {
		log.Fatal(err)
	}
	db, err := store.Open(filepath.Join(dataDir, "nexo.json"))
	if err != nil {
		log.Fatal(err)
	}
	api := httpapi.New(db, dataDir, webDir, os.Getenv("NEXO_ACCESS_TOKEN"), os.Getenv("NEXO_PAYMENT_URL"), os.Getenv("NEXO_PRO_KEY"), os.Getenv("NEXO_PRISMATEC_URL"), os.Getenv("NEXO_PRISMATEC_TOKEN"), os.Getenv("NEXO_MININODE_URL"))
	srv := &http.Server{Addr: addr, Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Minute, WriteTimeout: 30 * time.Minute, IdleTimeout: 90 * time.Second}
	log.Printf("NEXO 0.1 listening on %s", addr)
	log.Printf("data directory: %s", dataDir)
	log.Fatal(srv.ListenAndServe())
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
