package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/markmilligan/goth-hw/internal/config"
	"github.com/markmilligan/goth-hw/internal/handlers"
	"github.com/markmilligan/goth-hw/internal/salesforce"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	sfClient := salesforce.NewClient(cfg.SFClientID, cfg.SFClientSecret, cfg.SFLoginURL)
	h := handlers.New(sfClient)

	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	mux.HandleFunc("GET /", h.Home)
	mux.HandleFunc("GET /activities", h.ActivityPage)
	mux.HandleFunc("GET /activities/rows", h.ActivityRows)
	mux.HandleFunc("GET /objects/{type}", h.ObjectPage)
	mux.HandleFunc("GET /objects/{type}/rows", h.ObjectRows)

	addr := ":" + cfg.Port
	log.Printf("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
