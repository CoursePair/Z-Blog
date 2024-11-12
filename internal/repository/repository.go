package repository

import (
	config "Z-Blog/internal/configuration"
	"Z-Blog/internal/model"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
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
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id INT AUTO_INCREMENT PRIMARY KEY, username VARCHAR(50) UNIQUE NOT NULL, password VARCHAR(255) NOT NULL);")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS blogs (id CHAR(36) PRIMARY KEY, creation_date DATETIME DEFAULT CURRENT_TIMESTAMP, headline VARCHAR(255) NOT NULL, text TEXT NOT NULL, user_id INT, FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE);")
	if err != nil {
		panic(err)
	}
}

func SaveBlog(entry model.BlogEntry) error {
	_, err := db.Exec("INSERT INTO blogs (id, creation_date, headline, text, user_id) VALUES (?, ?, ?, ?, ?)", entry.ID, entry.CreationDate, entry.Headline, entry.Text, entry.UserId)
	return err
}

func ReadBlogEntries(userId int) ([]model.BlogEntry, error) {
	var blogs []model.BlogEntry
	if db == nil {
		return nil, fmt.Errorf("Datenbankverbindung ist nicht vorhanden")
	}

	log.Printf("User ID für die Abfrage: %s", userId)

	// Abfrage der Blogs, die zur angegebenen UUID gehören
	rows, err := db.Query("SELECT id, creation_date, headline, text, user_id FROM blogs WHERE user_id = ?", userId)
	if err != nil {
		log.Printf("Fehler bei der Datenabfrage: %v", err)
		return nil, fmt.Errorf("Fehler bei der Datenabfrage: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var blog model.BlogEntry
		err := rows.Scan(&blog.ID, &blog.CreationDate, &blog.Headline, &blog.Text, &blog.UserId)
		if err != nil {
			log.Printf("Fehler beim Scannen der Zeilen: %v", err)
			return nil, fmt.Errorf("Fehler beim Scannen der Zeilen: %v", err)
		}
		blogs = append(blogs, blog)
		count++
	}
	log.Printf("Anzahl der abgerufenen Blogs: %d", count)

	if err := rows.Err(); err != nil {
		log.Printf("Fehler beim Durchlaufen der Zeilen: %v", err)
		return nil, fmt.Errorf("Fehler beim Durchlaufen der Zeilen: %v", err)
	}

	return blogs, nil
}

func ReadSpecificBlog(id string, userId int) (*model.BlogEntry, error) {
	var blog model.BlogEntry
	if db == nil {
		return nil, fmt.Errorf("Datenbankverbindung ist nicht vorhanden")
	}
	row := db.QueryRow("SELECT * FROM blogs WHERE id= ? AND user_id= ?", id, userId)

	row.Scan(&blog.ID, &blog.CreationDate, &blog.Headline, &blog.Text, &blog.UserId)
	return &blog, nil

}

func DeleteSpecificBlog(id string, userId int) error {
	if db == nil {
		return fmt.Errorf("Datenbankverbindung ist nicht vorhanden")
	}
	result, err := db.Exec("DELETE FROM blogs WHERE id=? AND user_id=?", id, userId)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fehler beim Abrufen der Anzahl der aktualisierten Zeilen: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return err
}

func UpdateSpecificBlog(id string, entry model.BlogEntry, userId int) error {
	result, err := db.Exec("UPDATE blogs SET text=?, headline=? WHERE id=? AND user_id=?", entry.Text, entry.Headline, id, userId)
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

func CreateUser(user model.Credentials) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, hashedPassword)
	if err != nil {
		log.Printf("Fehler bei der Datenabfrage: %v", err)
	}
}

func CheckUsername(username string) (bool, error) {
	var existingUser string
	err := db.QueryRow("SELECT username FROM users WHERE username=?", username).Scan(&existingUser)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if existingUser != "" {
		return true, nil
	}
	return false, nil
}

func SearchForUser(username string) (int, string, error) {
	var userId int
	var storedHashedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE username=?", username).Scan(&userId, &storedHashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", fmt.Errorf("Fehler bei der Datenabfrage: %v", err)
		}
	}

	return userId, storedHashedPassword, nil
}
