package main

import (
	config "Z-Blog/internal/configuration"
	"Z-Blog/internal/handlers"
	"Z-Blog/internal/repository"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.Get()
	if err != nil {
		log.Fatal("Fehler beim laden der Konfig", err)
	}

	address := cfg.Adress
	repository.InitDB()
	http.HandleFunc("/blogs", handlers.HandleBlogs)
	http.HandleFunc("/blogs/", handlers.HandlerForSpecificBlog)

	err = http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println(err)
	}
}
