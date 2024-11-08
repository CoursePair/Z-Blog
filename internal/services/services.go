package services

import (
	"Z-Blog/internal/model"
	"Z-Blog/internal/repository"
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

func GetBlogEntries(w http.ResponseWriter) {
	blogs, err := repository.ReadBlogEntries()
	if err != nil {
		http.Error(w, "Fehler beim Lesen der Tweets", http.StatusInternalServerError)
		return
	}
	if blogs == nil {
		err := json.NewEncoder(w).Encode([]model.BlogEntry{})
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(blogs)
	if err != nil {
		http.Error(w, "Fehler beim Ausgeben des Blog-Eintrags", http.StatusInternalServerError)
		return
	}
}

func SaveBlogEntry(w http.ResponseWriter, r *http.Request) {
	var blogEntry model.BlogEntry
	err := json.NewDecoder(r.Body).Decode(&blogEntry)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	blogEntry.ID = uuid.New()
	blogEntry.CreationDate = time.Now()
	err = repository.SaveBlog(blogEntry)
	if err != nil {
		http.Error(w, "Fehler beim Speichern des Blog-Eintrags: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header()
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(blogEntry)
	if err != nil {
		http.Error(w, "Fehler beim Ausgeben des Blog-Eintrags", http.StatusInternalServerError)
	}
}

func SpecificBlog(w http.ResponseWriter, r *http.Request) {
	blogId := strings.TrimPrefix(r.URL.Path, "/blogs/")
	tweet, err := repository.ReadSpecificBlog(blogId)
	if _, err := uuid.Parse(blogId); err != nil {
		http.Error(w, "Ungültiges ID-Format", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Fehler beim Lesen der Tweets", http.StatusInternalServerError)
		return
	}
	if tweet.ID == uuid.Nil {
		http.Error(w, "Tweet nicht gefunden", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(tweet)
	if err != nil {
		http.Error(w, "Fehler beim Ausgeben des Blog-Eintrags", http.StatusInternalServerError)
		return
	}
}

func DeleteBlogEntry(w http.ResponseWriter, r *http.Request) {
	blogId := strings.TrimPrefix(r.URL.Path, "/blogs/")
	err := repository.DeleteSpecificBlog(blogId)
	if _, err := uuid.Parse(blogId); err != nil {
		http.Error(w, "Ungültiges ID-Format", http.StatusBadRequest)
		return
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Blog-Eintrag nicht gefunden", http.StatusNotFound)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func UpdateBlogEntry(w http.ResponseWriter, r *http.Request) {
	blogId := strings.TrimPrefix(r.URL.Path, "/blogs/")
	var blogEntry model.BlogEntry
	err := json.NewDecoder(r.Body).Decode(&blogEntry)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(blogId); err != nil {
		http.Error(w, "Ungültiges ID-Format", http.StatusBadRequest)
		return
	}
	err = repository.UpdateSpecificBlog(blogId, blogEntry)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Blog-Eintrag nicht gefunden", http.StatusNotFound)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
