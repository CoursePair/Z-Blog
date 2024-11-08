package handlers

import (
	"Z-Blog/internal/services"
	"net/http"
)

func HandleBlogs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		services.SaveBlogEntry(w, r)
	case http.MethodGet:
		services.GetBlogEntries(w)
	default:
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
	}
}

func HandlerForSpecificBlog(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		services.SpecificBlog(w, r)
	case http.MethodDelete:
		services.DeleteBlogEntry(w, r)
	case http.MethodPatch:
		services.UpdateBlogEntry(w, r)
	default:
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
	}
}
