package main

import (
	config "Z-Blog/internal/configuration"
	"Z-Blog/internal/handlers"
	"Z-Blog/internal/repository"
	"Z-Blog/internal/services"
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

	handleBlogs := http.HandlerFunc(handlers.HandleBlogs)
	handleSpecificBlog := http.HandlerFunc(handlers.HandlerForSpecificBlog)

	blogs := services.JWTMiddleware(handleBlogs)
	specificBlog := services.JWTMiddleware(handleSpecificBlog)

	repository.InitDB()
	http.HandleFunc("/register", handlers.HandlerUserRegister)
	http.HandleFunc("/login", handlers.HandlerUserLogin)
	http.Handle("/blogs", blogs)
	http.Handle("/blogs/", specificBlog)

	err = http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println(err)
	}
}
