package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mattc/badger/internal/circleci"
	"github.com/mattc/badger/internal/handler"
)

func main() {
	addr := envOrDefault("BADGER_ADDR", ":8080")
	token := os.Getenv("CIRCLECI_TOKEN")

	h := handler.New(circleci.NewClient(token))
	log.Printf("badger listening on %s", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
