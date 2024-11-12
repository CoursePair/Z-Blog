package services

import (
	"Z-Blog/internal/model"
	"Z-Blog/internal/repository"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

var secretKey = []byte("Key")

func GetBlogEntries(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userId").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}
	blogs, err := repository.ReadBlogEntries(userID)
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
	userID, err := getUserId(r, w)
	if err != nil {
		http.Error(w, "Fehler beim Aufrufen der UserId", http.StatusBadRequest)
	}
	blogEntry.UserId = userID

	blogEntry.ID = uuid.New()
	blogEntry.CreationDate = time.Now()
	err = repository.SaveBlog(blogEntry)
	if err != nil {
		http.Error(w, "Fehler beim Speichern des Blog-Eintrags: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(blogEntry)
	if err != nil {
		http.Error(w, "Fehler beim Ausgeben des Blog-Eintrags", http.StatusInternalServerError)
	}
}

func SpecificBlog(w http.ResponseWriter, r *http.Request) {
	blogId := strings.TrimPrefix(r.URL.Path, "/blogs/")
	userID, err := getUserId(r, w)
	if err != nil {
		http.Error(w, "Fehler beim Aufrufen der UserId", http.StatusBadRequest)
	}
	blog, err := repository.ReadSpecificBlog(blogId, userID)
	if _, err := uuid.Parse(blogId); err != nil {
		http.Error(w, "Ungültiges ID-Format", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Fehler beim Lesen der Tweets", http.StatusInternalServerError)
		return
	}
	if blog.ID == uuid.Nil {
		http.Error(w, "Tweet nicht gefunden", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(blog)
	if err != nil {
		http.Error(w, "Fehler beim Ausgeben des Blog-Eintrags", http.StatusInternalServerError)
		return
	}
}

func DeleteBlogEntry(w http.ResponseWriter, r *http.Request) {
	blogId := strings.TrimPrefix(r.URL.Path, "/blogs/")
	userID, err := getUserId(r, w)
	if err != nil {
		http.Error(w, "Fehler beim Aufrufen der UserId", http.StatusBadRequest)
	}
	err = repository.DeleteSpecificBlog(blogId, userID)
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
	userID, err := getUserId(r, w)
	if err != nil {
		http.Error(w, "Fehler beim Aufrufen der UserId", http.StatusBadRequest)
	}
	var blogEntry model.BlogEntry
	err = json.NewDecoder(r.Body).Decode(&blogEntry)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(blogId); err != nil {
		http.Error(w, "Ungültiges ID-Format", http.StatusBadRequest)
		return
	}
	err = repository.UpdateSpecificBlog(blogId, blogEntry, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Blog-Eintrag nicht gefunden", http.StatusNotFound)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	var cred model.Credentials
	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
	}

	isTaken, err := repository.CheckUsername(cred.Username)
	if err != nil {
		http.Error(w, "Datenbankfehler", http.StatusInternalServerError)
		return
	}
	if isTaken {
		http.Error(w, "Benutzername ist bereits vergeben", http.StatusConflict)
		return
	}
	repository.CreateUser(cred)
	w.Write([]byte("Registriert"))
}

func Login(w http.ResponseWriter, r *http.Request) {
	var cred model.Credentials
	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
	}

	userId, storedHashPassword, err := repository.SearchForUser(cred.Username)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHashPassword), []byte(cred.Password)); err != nil {
		http.Error(w, "Ungültige Anmeldeinformationen", http.StatusUnauthorized)
		return
	}

	tokenString, err := createJWT(userId)
	if err != nil {
		http.Error(w, "Fehler beim Erstellen des Tokens", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", "Bearer "+tokenString)
}

func createJWT(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Kein gültiges Token vorhanden", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unerwartete Signaturmethode: %v", token.Header["alg"])
			}
			return secretKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Ungültiges Token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Ungültiges Token", http.StatusUnauthorized)
			return
		}
		userId, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Benutzer-ID nicht gefunden im Token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userId", int(userId))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserId(r *http.Request, w http.ResponseWriter) (int, error) {
	userID, ok := r.Context().Value("userId").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}
	return userID, nil
}
