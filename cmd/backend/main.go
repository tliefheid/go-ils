package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/yourusername/library-ils-backend/backend"
)

func main() {
	cfg := backend.LoadConfig()
	db := backend.InitDB(cfg.DSN())

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing DB: %v", err)
		}
	}()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, "OK"); err != nil {
			log.Printf("Error writing health response: %v", err)
		}
	})

	http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listBooksHandler(db, w, r)
		case http.MethodPost:
			addBookHandler(db, w, r)
		case http.MethodPut:
			editBookHandler(db, w, r)
		case http.MethodDelete:
			deleteBookHandler(db, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/books/search", func(w http.ResponseWriter, r *http.Request) {
		searchBooksHandler(db, w, r)
	})

	http.HandleFunc("/members", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listMembersHandler(db, w, r)
		case http.MethodPost:
			addMemberHandler(db, w, r)
		case http.MethodPut:
			editMemberHandler(db, w, r)
		case http.MethodDelete:
			deleteMemberHandler(db, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/borrow", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			borrowBookHandler(db, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/return", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			returnBookHandler(db, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/reports/borrowed", func(w http.ResponseWriter, r *http.Request) {
		borrowedBooksReportHandler(db, w, r)
	})

	http.HandleFunc("/reports/overdue", func(w http.ResponseWriter, r *http.Request) {
		overdueBooksReportHandler(db, w, r)
	})
	http.HandleFunc("/isbn", isbnInfoHandler)

	http.HandleFunc("/reports/member", func(w http.ResponseWriter, r *http.Request) {
		memberHistoryReportHandler(db, w, r)
	})

	// Add /borrowing endpoint to serve borrowing details by id
	http.HandleFunc("/borrowing", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			http.Error(w, "Missing borrowing id", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil || id == 0 {
			http.Error(w, "Invalid borrowing id", http.StatusBadRequest)
			return
		}

		var detail struct {
			ID         int        `json:"id"`
			BookID     int        `json:"book_id"`
			BookTitle  string     `json:"book_title"`
			MemberID   int        `json:"member_id"`
			MemberName string     `json:"member_name"`
			IssueDate  time.Time  `json:"issue_date"`
			DueDate    time.Time  `json:"due_date"`
			ReturnDate *time.Time `json:"return_date"`
			Fine       float64    `json:"fine"`
		}

		row := db.QueryRow(`SELECT br.id, br.book_id, b.title, br.member_id, m.name, br.issue_date, br.due_date, br.return_date, br.fine FROM borrowings br JOIN books b ON br.book_id = b.id JOIN members m ON br.member_id = m.id WHERE br.id = $1`, id)

		err = row.Scan(&detail.ID, &detail.BookID, &detail.BookTitle, &detail.MemberID, &detail.MemberName, &detail.IssueDate, &detail.DueDate, &detail.ReturnDate, &detail.Fine)
		if err != nil {
			http.Error(w, "Borrowing not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(detail)
	})

	// Add /borrowed-detail endpoint to get borrowing by book_id and member (unreturned)
	http.HandleFunc("/borrowed-detail", func(w http.ResponseWriter, r *http.Request) {
		bookIDStr := r.URL.Query().Get("book_id")
		memberIDStr := r.URL.Query().Get("member_id")

		if bookIDStr == "" || memberIDStr == "" {
			http.Error(w, "Missing book_id or member_id", http.StatusBadRequest)
			return
		}

		bookID, err := strconv.Atoi(bookIDStr)
		if err != nil || bookID == 0 {
			http.Error(w, "Invalid book_id", http.StatusBadRequest)
			return
		}

		memberID, err := strconv.Atoi(memberIDStr)
		if err != nil || memberID == 0 {
			http.Error(w, "Invalid member_id", http.StatusBadRequest)
			return
		}

		var detail struct {
			ID         int        `json:"id"`
			BookID     int        `json:"book_id"`
			BookTitle  string     `json:"book_title"`
			MemberID   int        `json:"member_id"`
			MemberName string     `json:"member_name"`
			IssueDate  time.Time  `json:"issue_date"`
			DueDate    time.Time  `json:"due_date"`
			ReturnDate *time.Time `json:"return_date"`
			Fine       float64    `json:"fine"`
		}

		row := db.QueryRow(`SELECT br.id, br.book_id, b.title, br.member_id, m.name, br.issue_date, br.due_date, br.return_date, br.fine FROM borrowings br JOIN books b ON br.book_id = b.id JOIN members m ON br.member_id = m.id WHERE br.book_id = $1 AND br.member_id = $2 AND br.return_date IS NULL ORDER BY br.issue_date DESC LIMIT 1`, bookID, memberID)

		err = row.Scan(&detail.ID, &detail.BookID, &detail.BookTitle, &detail.MemberID, &detail.MemberName, &detail.IssueDate, &detail.DueDate, &detail.ReturnDate, &detail.Fine)
		if err != nil {
			http.Error(w, "Borrowing not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(detail)
	})

	fmt.Println("Library ILS Backend - Go API running on :8180")

	if err := http.ListenAndServe(":8180", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// isbnInfoHandler fetches book info from Open Library API by ISBN
func isbnInfoHandler(w http.ResponseWriter, r *http.Request) {
	isbn := r.URL.Query().Get("isbn")
	if isbn == "" {
		http.Error(w, "Missing isbn parameter", http.StatusBadRequest)
		return
	}

	book, err := backend.LookupByISBN(isbn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching book info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func runMigrations(db *sql.DB) error {
	data, err := os.ReadFile("migrations.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(data))

	return err
}

// --- Book Handlers ---
func listBooksHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, author, isbn, genre, publication_year, copies_total, copies_available FROM books")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []backend.Book

	for rows.Next() {
		var b backend.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Genre, &b.PublicationYear, &b.CopiesTotal, &b.CopiesAvailable); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		books = append(books, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func addBookHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var b backend.Book

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &b); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO books (title, author, isbn, genre, publication_year, copies_total, copies_available) VALUES ($1, $2, $3, $4, $5, $6, $6) RETURNING id`

	err = db.QueryRow(query, b.Title, b.Author, b.ISBN, b.Genre, b.PublicationYear, b.CopiesTotal).Scan(&b.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func editBookHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var b backend.Book

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &b); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if b.ID == 0 {
		http.Error(w, "Missing book ID", http.StatusBadRequest)
		return
	}

	query := `UPDATE books SET title=$1, author=$2, isbn=$3, genre=$4, publication_year=$5, copies_total=$6, copies_available=$7 WHERE id=$8`

	_, err = db.Exec(query, b.Title, b.Author, b.ISBN, b.Genre, b.PublicationYear, b.CopiesTotal, b.CopiesAvailable, b.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteBookHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM books WHERE id=$1", id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func searchBooksHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	rows, err := db.Query(`SELECT id, title, author, isbn, genre, publication_year, copies_total, copies_available FROM books WHERE title ILIKE '%' || $1 || '%' OR author ILIKE '%' || $1 || '%' OR isbn ILIKE '%' || $1 || '%'`, q)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var books []backend.Book

	for rows.Next() {
		var b backend.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Genre, &b.PublicationYear, &b.CopiesTotal, &b.CopiesAvailable); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		books = append(books, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// --- Member Handlers ---
func listMembersHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, contact, member_id FROM members")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var members []backend.Member

	for rows.Next() {
		var m backend.Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Contact, &m.MemberID); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		members = append(members, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func addMemberHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var m backend.Member

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &m); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO members (name, contact, member_id) VALUES ($1, $2, $3) RETURNING id`

	err = db.QueryRow(query, m.Name, m.Contact, m.MemberID).Scan(&m.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func editMemberHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var m backend.Member

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &m); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if m.ID == 0 {
		http.Error(w, "Missing member ID", http.StatusBadRequest)
		return
	}

	query := `UPDATE members SET name=$1, contact=$2, member_id=$3 WHERE id=$4`

	_, err = db.Exec(query, m.Name, m.Contact, m.MemberID, m.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteMemberHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM members WHERE id=$1", id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Borrow/Return Handlers ---
func borrowBookHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookID   string `json:"book_id"`
		MemberID string `json:"member_id"`
		Days     string `json:"days"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Printf("string(body): %v\n", string(body))

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.BookID == "" || req.MemberID == "" || req.Days <= "" {
		http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
		return
	}
	// Check book availability
	var available int

	bookID, err := strconv.Atoi(req.BookID)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	err = db.QueryRow("SELECT copies_available FROM books WHERE id=$1", bookID).Scan(&available)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	if available < 1 {
		http.Error(w, "No copies available", http.StatusConflict)
		return
	}
	// Insert borrowing record
	issueDate := time.Now()

	days, err := strconv.Atoi(req.Days)
	if err != nil || days <= 0 {
		http.Error(w, "Invalid number of days", http.StatusBadRequest)
		return
	}

	dueDate := issueDate.AddDate(0, 0, days)

	memberID, err := strconv.Atoi(req.MemberID)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`INSERT INTO borrowings (book_id, member_id, issue_date, due_date) VALUES ($1, $2, $3, $4)`, bookID, memberID, issueDate, dueDate)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	// Update book inventory
	_, err = db.Exec("UPDATE books SET copies_available = copies_available - 1 WHERE id=$1", bookID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func returnBookHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookID   int `json:"book_id"`
		MemberID int `json:"member_id"`
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.BookID == 0 || req.MemberID == 0 {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}
	// Find the latest unreturned borrowing
	var borrowID int

	var dueDate time.Time

	err = db.QueryRow(`SELECT id, due_date FROM borrowings WHERE book_id=$1 AND member_id=$2 AND return_date IS NULL ORDER BY issue_date DESC LIMIT 1`, req.BookID, req.MemberID).Scan(&borrowID, &dueDate)
	if err != nil {
		http.Error(w, "No active borrowing found", http.StatusNotFound)
		return
	}

	returnDate := time.Now()
	fine := 0.0

	if returnDate.After(dueDate) {
		daysLate := int(returnDate.Sub(dueDate).Hours() / 24)
		if daysLate > 0 {
			fine = float64(daysLate) * 1.0 // $1 per day late
		}
	}

	_, err = db.Exec(`UPDATE borrowings SET return_date=$1, fine=$2 WHERE id=$3`, returnDate, fine, borrowID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE books SET copies_available = copies_available + 1 WHERE id=$1", req.BookID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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
