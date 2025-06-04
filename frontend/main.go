package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
)

const backendURL = "http://localhost:8180"

var templates = template.Must(template.New("").ParseGlob("templates/*.gohtml"))

type Book struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	Author          string `json:"author"`
	ISBN            string `json:"isbn"`
	Genre           string `json:"genre"`
	PublicationYear int    `json:"publication_year"`
	CopiesTotal     int    `json:"copies_total"`
	CopiesAvailable int    `json:"copies_available"`
}

type Member struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Contact  string `json:"contact"`
	MemberID string `json:"member_id"`
}

type BorrowedBook struct {
	BookID    int    `json:"book_id"`
	Title     string `json:"title"`
	Member    string `json:"member"`
	IssueDate string `json:"issue_date"`
	DueDate   string `json:"due_date"`
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", booksPage)
	http.HandleFunc("/book", bookDetailPage)
	http.HandleFunc("/members", membersPage)
	http.HandleFunc("/borrowed", borrowedBooksPage)
	http.HandleFunc("/borrow", borrowBookHandler)
	log.Println("Frontend UI running on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func booksPage(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(backendURL + "/books")
	if err != nil {
		http.Error(w, "Failed to fetch books", 500)
		return
	}
	defer resp.Body.Close()
	var books []Book
	if err := json.NewDecoder(resp.Body).Decode(&books); err != nil {
		http.Error(w, "Failed to decode books", 500)
		return
	}
	templates.ExecuteTemplate(w, "books.gohtml", books)
}

func bookDetailPage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing book id", 400)
		return
	}
	resp, err := http.Get(backendURL + "/books")
	if err != nil {
		http.Error(w, "Failed to fetch books", 500)
		return
	}
	defer resp.Body.Close()
	var books []Book
	if err := json.NewDecoder(resp.Body).Decode(&books); err != nil {
		http.Error(w, "Failed to decode books", 500)
		return
	}
	var book *Book
	for _, b := range books {
		if fmt.Sprintf("%d", b.ID) == id {
			book = &b
			break
		}
	}
	if book == nil {
		http.Error(w, "Book not found", 404)
		return
	}
	// Fetch members for borrow dropdown
	resp2, err := http.Get(backendURL + "/members")
	if err != nil {
		http.Error(w, "Failed to fetch members", 500)
		return
	}
	defer resp2.Body.Close()
	var members []Member
	if err := json.NewDecoder(resp2.Body).Decode(&members); err != nil {
		http.Error(w, "Failed to decode members", 500)
		return
	}
	templates.ExecuteTemplate(w, "book_detail.gohtml", struct {
		Book    *Book
		Members []Member
	}{book, members})
}

func membersPage(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(backendURL + "/members")
	if err != nil {
		http.Error(w, "Failed to fetch members", 500)
		return
	}
	defer resp.Body.Close()
	var members []Member
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		http.Error(w, "Failed to decode members", 500)
		return
	}
	templates.ExecuteTemplate(w, "members.gohtml", members)
}

func borrowedBooksPage(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(backendURL + "/reports/borrowed")
	if err != nil {
		http.Error(w, "Failed to fetch borrowed books", 500)
		return
	}
	defer resp.Body.Close()
	var borrowed []BorrowedBook
	if err := json.NewDecoder(resp.Body).Decode(&borrowed); err != nil {
		http.Error(w, "Failed to decode borrowed books", 500)
		return
	}
	templates.ExecuteTemplate(w, "borrowed.gohtml", borrowed)
}

func borrowBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	bookID := r.FormValue("book_id")
	memberID := r.FormValue("member_id")
	days := r.FormValue("days")
	if bookID == "" || memberID == "" || days == "" {
		http.Error(w, "Missing fields", 400)
		return
	}
	payload := map[string]string{
		"book_id":   bookID,
		"member_id": memberID,
		"days":      days,
	}
	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post(backendURL+"/borrow", "application/json", bytes.NewReader(jsonPayload))
	if err != nil {
		http.Error(w, "Failed to borrow book", 500)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, string(body), resp.StatusCode)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
