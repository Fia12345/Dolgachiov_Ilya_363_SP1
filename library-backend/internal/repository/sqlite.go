package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"library-backend/internal/models"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.initTables(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteRepository) Close() {
	r.db.Close()
}

func (r *SQLiteRepository) initTables() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS books (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			author TEXT NOT NULL,
			isbn TEXT UNIQUE,
			year INTEGER,
			status TEXT DEFAULT 'Available'
		);

		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			registration_date DATETIME DEFAULT CURRENT_TIMESTAMP,
			password TEXT NOT NULL,
			role TEXT DEFAULT 'user'
		);

		CREATE TABLE IF NOT EXISTS issues (
			id TEXT PRIMARY KEY,
			book_id TEXT REFERENCES books(id),
			user_id TEXT REFERENCES users(id),
			issue_date DATETIME DEFAULT CURRENT_TIMESTAMP,
			due_date DATETIME,
			return_date DATETIME
		);
	`)
	return err
}

func (r *SQLiteRepository) CreateBook(book *models.Book) error {
	_, err := r.db.Exec("INSERT INTO books (id, title, author, isbn, year, status) VALUES (?, ?, ?, ?, ?, ?)",
		book.ID, book.Title, book.Author, book.ISBN, book.Year, book.Status)
	return err
}

func (r *SQLiteRepository) GetBook(id string) (*models.Book, error) {
	var b models.Book
	err := r.db.QueryRow("SELECT id, title, author, isbn, year, status FROM books WHERE id = ?", id).
		Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("book not found")
	}
	return &b, err
}

func (r *SQLiteRepository) GetBooks(page, limit int, author, status string) ([]models.Book, error) {
	query := "SELECT id, title, author, isbn, year, status FROM books WHERE 1=1"
	args := []interface{}{}

	if author != "" {
		query += " AND author LIKE ?"
		args = append(args, "%"+author+"%")
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, (page-1)*limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Status)
		books = append(books, b)
	}
	return books, nil
}

func (r *SQLiteRepository) UpdateBook(book *models.Book) error {
	_, err := r.db.Exec("UPDATE books SET title=?, author=?, isbn=?, year=?, status=? WHERE id=?",
		book.Title, book.Author, book.ISBN, book.Year, book.Status, book.ID)
	return err
}

func (r *SQLiteRepository) CreateUser(user *models.User, password, role string) error {
	_, err := r.db.Exec("INSERT INTO users (id, name, email, password, role) VALUES (?, ?, ?, ?, ?)",
		user.ID, user.Name, user.Email, password, role)
	return err
}

func (r *SQLiteRepository) GetUserByEmail(email string) (*models.User, string, string, error) {
	var u models.User
	var pass, role string
	err := r.db.QueryRow("SELECT id, name, email, registration_date, password, role FROM users WHERE email = ?", email).
		Scan(&u.ID, &u.Name, &u.Email, &u.RegistrationDate, &pass, &role)
	return &u, pass, role, err
}

func (r *SQLiteRepository) GetUserBooks(userID string) ([]models.Book, error) {
	rows, err := r.db.Query(`
		SELECT b.id, b.title, b.author, b.isbn, b.year, b.status 
		FROM books b JOIN issues i ON b.id = i.book_id 
		WHERE i.user_id = ? AND i.return_date IS NULL`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Status)
		books = append(books, b)
	}
	return books, nil
}

func (r *SQLiteRepository) IssueBook(bookID, userID string) error {
	var status string
	err := r.db.QueryRow("SELECT status FROM books WHERE id = ?", bookID).Scan(&status)
	if err != nil || status != "Available" {
		return fmt.Errorf("book not available")
	}

	issueID := "issue_" + bookID + "_" + userID 
	dueDate := time.Now().Add(14 * 24 * time.Hour)

	_, err = r.db.Exec("INSERT INTO issues (id, book_id, user_id, due_date) VALUES (?, ?, ?, ?)",
		issueID, bookID, userID, dueDate)
	if err != nil {
		return err
	}

	_, err = r.db.Exec("UPDATE books SET status = 'Issued' WHERE id = ?", bookID)
	return err
}

func (r *SQLiteRepository) ReturnBook(bookID string) error {
	_, err := r.db.Exec("UPDATE issues SET return_date = CURRENT_TIMESTAMP WHERE book_id = ? AND return_date IS NULL", bookID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec("UPDATE books SET status = 'Available' WHERE id = ?", bookID)
	return err
}
