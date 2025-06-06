package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/yourusername/library-ils-backend/internal/backend"
	"github.com/yourusername/library-ils-backend/internal/repository/postgres"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	HTTP       string
}

func LoadConfig() Config {
	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "library"),
		HTTP:       ":8182",
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := LoadConfig()

	db, err := postgres.NewStore(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing DB: %v", err)
		}
	}()

	s, err := backend.New(backend.Config{
		Repository: db,
	})
	if err != nil {
		log.Fatalf("Failed to initialize backend service: %v", err)
	}

	err = db.Migrate("migrations.sql")
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	// 	if _, err := fmt.Fprintln(w, "OK"); err != nil {
	// 		log.Printf("Error writing health response: %v", err)
	// 	}
	// })

	// http.HandleFunc("/reports/borrowed", func(w http.ResponseWriter, r *http.Request) {
	// 	borrowedBooksReportHandler(db, w, r)
	// })

	// http.HandleFunc("/reports/overdue", func(w http.ResponseWriter, r *http.Request) {
	// 	overdueBooksReportHandler(db, w, r)
	// })

	// http.HandleFunc("/reports/member", func(w http.ResponseWriter, r *http.Request) {
	// 	memberHistoryReportHandler(db, w, r)
	// })

	// // Add /borrowing endpoint to serve borrowing details by id
	// http.HandleFunc("/borrowing", func(w http.ResponseWriter, r *http.Request) {
	// 	idStr := r.URL.Query().Get("id")
	// 	if idStr == "" {
	// 		http.Error(w, "Missing borrowing id", http.StatusBadRequest)
	// 		return
	// 	}

	// 	id, err := strconv.Atoi(idStr)
	// 	if err != nil || id == 0 {
	// 		http.Error(w, "Invalid borrowing id", http.StatusBadRequest)
	// 		return
	// 	}

	// 	var detail struct {
	// 		ID         int        `json:"id"`
	// 		BookID     int        `json:"book_id"`
	// 		BookTitle  string     `json:"book_title"`
	// 		MemberID   int        `json:"member_id"`
	// 		MemberName string     `json:"member_name"`
	// 		IssueDate  time.Time  `json:"issue_date"`
	// 		DueDate    time.Time  `json:"due_date"`
	// 		ReturnDate *time.Time `json:"return_date"`
	// 		Fine       float64    `json:"fine"`
	// 	}

	// 	row := db.QueryRow(`SELECT br.id, br.book_id, b.title, br.member_id, m.name, br.issue_date, br.due_date, br.return_date, br.fine FROM borrowings br JOIN books b ON br.book_id = b.id JOIN members m ON br.member_id = m.id WHERE br.id = $1`, id)

	// 	err = row.Scan(&detail.ID, &detail.BookID, &detail.BookTitle, &detail.MemberID, &detail.MemberName, &detail.IssueDate, &detail.DueDate, &detail.ReturnDate, &detail.Fine)
	// 	if err != nil {
	// 		http.Error(w, "Borrowing not found", http.StatusNotFound)
	// 		return
	// 	}

	// 	w.Header().Set("Content-Type", "application/json")
	// 	json.NewEncoder(w).Encode(detail)
	// })

	// // Add /borrowed-detail endpoint to get borrowing by book_id and member (unreturned)
	// http.HandleFunc("/borrowed-detail", func(w http.ResponseWriter, r *http.Request) {
	// 	bookIDStr := r.URL.Query().Get("book_id")
	// 	memberIDStr := r.URL.Query().Get("member_id")

	// 	if bookIDStr == "" || memberIDStr == "" {
	// 		http.Error(w, "Missing book_id or member_id", http.StatusBadRequest)
	// 		return
	// 	}

	// 	bookID, err := strconv.Atoi(bookIDStr)
	// 	if err != nil || bookID == 0 {
	// 		http.Error(w, "Invalid book_id", http.StatusBadRequest)
	// 		return
	// 	}

	// 	memberID, err := strconv.Atoi(memberIDStr)
	// 	if err != nil || memberID == 0 {
	// 		http.Error(w, "Invalid member_id", http.StatusBadRequest)
	// 		return
	// 	}

	// 	var detail struct {
	// 		ID         int        `json:"id"`
	// 		BookID     int        `json:"book_id"`
	// 		BookTitle  string     `json:"book_title"`
	// 		MemberID   int        `json:"member_id"`
	// 		MemberName string     `json:"member_name"`
	// 		IssueDate  time.Time  `json:"issue_date"`
	// 		DueDate    time.Time  `json:"due_date"`
	// 		ReturnDate *time.Time `json:"return_date"`
	// 		Fine       float64    `json:"fine"`
	// 	}

	// 	row := db.QueryRow(`SELECT br.id, br.book_id, b.title, br.member_id, m.name, br.issue_date, br.due_date, br.return_date, br.fine FROM borrowings br JOIN books b ON br.book_id = b.id JOIN members m ON br.member_id = m.id WHERE br.book_id = $1 AND br.member_id = $2 AND br.return_date IS NULL ORDER BY br.issue_date DESC LIMIT 1`, bookID, memberID)

	// 	err = row.Scan(&detail.ID, &detail.BookID, &detail.BookTitle, &detail.MemberID, &detail.MemberName, &detail.IssueDate, &detail.DueDate, &detail.ReturnDate, &detail.Fine)
	// 	if err != nil {
	// 		http.Error(w, "Borrowing not found", http.StatusNotFound)
	// 		return
	// 	}

	// 	w.Header().Set("Content-Type", "application/json")
	// 	json.NewEncoder(w).Encode(detail)
	// })

	fmt.Println("Library ILS Backend - Go API running on ", cfg.HTTP)

	srv := &http.Server{
		Addr:    cfg.HTTP,
		Handler: s.Mux(),
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.SetKeepAlivesEnabled(false)

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	go func() {
		fmt.Println("Starting server on", cfg.HTTP)

		if err := http.ListenAndServe(cfg.HTTP, srv.Handler); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	fmt.Println("wait for shutdown signal...")
	<-ctx.Done()
	fmt.Println("shutting down...")
}

// func searchBooksHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
// 	q := r.URL.Query().Get("q")

// 	rows, err := db.Query(`SELECT id, title, author, isbn, publication_year, copies_total, copies_available FROM books WHERE title ILIKE '%' || $1 || '%' OR author ILIKE '%' || $1 || '%' OR isbn ILIKE '%' || $1 || '%'`, q)
// 	if err != nil {
// 		http.Error(w, "Database error", http.StatusInternalServerError)
// 		return
// 	}

// 	defer rows.Close()

// 	var books []backend.Book

// 	for rows.Next() {
// 		var b backend.Book
// 		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.PublicationYear, &b.CopiesTotal, &b.CopiesAvailable); err != nil {
// 			http.Error(w, "Database error", http.StatusInternalServerError)
// 			return
// 		}

// 		books = append(books, b)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(books)
// }

// --- Borrow/Return Handlers ---

// --- Reporting Handlers ---
func borrowedBooksReportHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT b.id, b.title, m.name, br.issue_date, br.due_date FROM borrowings br JOIN books b ON br.book_id = b.id JOIN members m ON br.member_id = m.id WHERE br.return_date IS NULL`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var result []map[string]interface{}

	for rows.Next() {
		var id int

		var title, name string

		var issueDate, dueDate time.Time
		if err := rows.Scan(&id, &title, &name, &issueDate, &dueDate); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		result = append(result, map[string]interface{}{
			"book_id":    id,
			"title":      title,
			"member":     name,
			"issue_date": issueDate,
			"due_date":   dueDate,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func overdueBooksReportHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT b.id, b.title, m.name, br.issue_date, br.due_date FROM borrowings br JOIN books b ON br.book_id = b.id JOIN members m ON br.member_id = m.id WHERE br.return_date IS NULL AND br.due_date < NOW()`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var result []map[string]interface{}

	for rows.Next() {
		var id int

		var title, name string

		var issueDate, dueDate time.Time
		if err := rows.Scan(&id, &title, &name, &issueDate, &dueDate); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		result = append(result, map[string]interface{}{
			"book_id":    id,
			"title":      title,
			"member":     name,
			"issue_date": issueDate,
			"due_date":   dueDate,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func memberHistoryReportHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	memberIDStr := r.URL.Query().Get("member_id")

	memberID, err := strconv.Atoi(memberIDStr)
	if err != nil || memberID == 0 {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`SELECT b.title, br.issue_date, br.due_date, br.return_date, br.fine FROM borrowings br JOIN books b ON br.book_id = b.id WHERE br.member_id = $1`, memberID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var result []map[string]interface{}

	for rows.Next() {
		var title string

		var issueDate, dueDate time.Time

		var returnDate sql.NullTime

		var fine float64
		if err := rows.Scan(&title, &issueDate, &dueDate, &returnDate, &fine); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		var retDate interface{}
		if returnDate.Valid {
			retDate = returnDate.Time
		} else {
			retDate = nil
		}

		result = append(result, map[string]interface{}{
			"title":       title,
			"issue_date":  issueDate,
			"due_date":    dueDate,
			"return_date": retDate,
			"fine":        fine,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
