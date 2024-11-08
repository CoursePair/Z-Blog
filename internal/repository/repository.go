package repository

import (
	config "Z-Blog/internal/configuration"
	"Z-Blog/internal/model"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func InitDB() {
	var err error
	cnf, err := config.Get()
	if err != nil {
		log.Fatal(err)
	}
	dsn := fmt.Sprintf("user:password@tcp(%s:%d)/maxim_db?parseTime=true", cnf.Host, cnf.Port)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS blogs (id CHAR(36) PRIMARY KEY, creation_date DATETIME DEFAULT CURRENT_TIMESTAMP, headline VARCHAR(255) NOT NULL, text TEXT NOT NULL);")
	if err != nil {
		panic(err)
	}
}

func SaveBlog(entry model.BlogEntry) error {
	_, err := db.Exec("INSERT INTO blogs (id, creation_date, headline, text) VALUES (?, ?, ?, ?)", entry.ID, entry.CreationDate, entry.Headline, entry.Text)
	return err
}

func ReadBlogEntries() ([]model.BlogEntry, error) {
	var blogs []model.BlogEntry
	if db == nil {
		return nil, fmt.Errorf("Datenbankverbindung ist nicht vorhanden")
	}
	rows, err := db.Query("SELECT * FROM blogs")
	if err != nil {
		log.Printf("Fehler bei der Datenabfrage: %v", err)
		return nil, fmt.Errorf("Fehler bei der Datenabfrage: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var blog model.BlogEntry
		if err := rows.Scan(&blog.ID, &blog.CreationDate, &blog.Headline, &blog.Text); err != nil {
			log.Printf("Fehler beim Scannen der Zeilen: %v", err)
			return nil, fmt.Errorf("Fehler beim Scannen der Zeilen: %v", err)
		}
		blogs = append(blogs, blog)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Fehler beim Durchlaufen der Zeilen: %v", err)
		return nil, fmt.Errorf("Fehler beim Durchlaufen der Zeilen: %v", err)
	}

	return blogs, nil
}

func ReadSpecificBlog(id string) (*model.BlogEntry, error) {
	var blog model.BlogEntry
	if db == nil {
		return nil, fmt.Errorf("Datenbankverbindung ist nicht vorhanden")
	}
	row := db.QueryRow("SELECT * FROM blogs WHERE id=?", id)

	row.Scan(&blog.ID, &blog.CreationDate, &blog.Headline, &blog.Text)
	return &blog, nil

}

func DeleteSpecificBlog(id string) error {
	if db == nil {
		return fmt.Errorf("Datenbankverbindung ist nicht vorhanden")
	}
	result, err := db.Exec("DELETE FROM blogs WHERE id=?", id)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fehler beim Abrufen der Anzahl der aktualisierten Zeilen: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return err
}

func UpdateSpecificBlog(id string, entry model.BlogEntry) error {
	result, err := db.Exec("UPDATE blogs SET text=?, headline=? WHERE id=?", entry.Text, entry.Headline, id)
	if err != nil {
		return fmt.Errorf("Fehler beim Ausführen des SQL-Updates: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fehler beim Abrufen der Anzahl der aktualisierten Zeilen: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
