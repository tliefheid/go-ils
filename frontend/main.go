package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
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

type BorrowingDetail struct {
	ID         int     `json:"id"`
	BookID     int     `json:"book_id"`
	BookTitle  string  `json:"book_title"`
	MemberID   int     `json:"member_id"`
	MemberName string  `json:"member_name"`
	IssueDate  string  `json:"issue_date"`
	DueDate    string  `json:"due_date"`
	ReturnDate *string `json:"return_date"`
	Fine       float64 `json:"fine"`
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", booksPage)
	http.HandleFunc("/book", bookDetailPage)
	http.HandleFunc("/members", membersPage)
	http.HandleFunc("/borrowed", borrowedBooksPage)
	http.HandleFunc("/borrow", borrowBookHandler)
	http.HandleFunc("/isbn-lookup", isbnLookupPage)
	http.HandleFunc("/delete-book", deleteBookHandler)
	http.HandleFunc("/update-book", updateBookHandler)
	http.HandleFunc("/borrowed-detail", borrowedDetailPage)
	http.HandleFunc("/return-borrowing", returnBorrowingHandler)
	http.HandleFunc("/reports", reportsPage)
	http.HandleFunc("/member", memberDetailPage)
	http.HandleFunc("/update-member", updateMemberHandler)
	http.HandleFunc("/delete-member", deleteMemberHandler)
	log.Println("Frontend UI running on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func booksPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var resp *http.Response

	var err error
	if q != "" {
		resp, err = http.Get(backendURL + "/books/search?q=" + q)
	} else {
		resp, err = http.Get(backendURL + "/books")
	}

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

	templates.ExecuteTemplate(w, "books.gohtml", map[string]interface{}{
		"Books": books,
		"Query": q,
	})
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
		if strconv.Itoa(b.ID) == id {
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

// Handler for borrowed book detail page
func borrowedDetailPage(w http.ResponseWriter, r *http.Request) {
	bookID := r.URL.Query().Get("book_id")
	memberName := r.URL.Query().Get("member")

	if bookID == "" || memberName == "" {
		http.Error(w, "Missing book_id or member", 400)
		return
	}
	// Fetch all members to resolve memberName to memberID
	respMembers, err := http.Get(backendURL + "/members")
	if err != nil {
		http.Error(w, "Failed to fetch members", 500)
		return
	}

	defer respMembers.Body.Close()

	var members []Member
	if err := json.NewDecoder(respMembers.Body).Decode(&members); err != nil {
		http.Error(w, "Failed to decode members", 500)
		return
	}

	var memberID string

	for _, m := range members {
		if m.Name == memberName {
			memberID = strconv.Itoa(m.ID)
			break
		}
	}

	if memberID == "" {
		http.Error(w, "Member not found", 404)
		return
	}
	// Now call backend with book_id and member_id
	resp, err := http.Get(backendURL + "/borrowed-detail?book_id=" + bookID + "&member_id=" + memberID)
	if err != nil {
		http.Error(w, "Failed to fetch borrowing detail", 500)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, string(body), resp.StatusCode)

		return
	}

	var detail BorrowingDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		http.Error(w, "Failed to decode borrowing detail", 500)
		return
	}

	templates.ExecuteTemplate(w, "borrowed_detail.gohtml", map[string]interface{}{"Borrowing": detail})
}

// Handler for returning a borrowed book from the detail page
func returnBorrowingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	borrowingID := r.FormValue("borrowing_id")
	if borrowingID == "" {
		http.Error(w, "Missing borrowing ID", 400)
		return
	}
	// Fetch borrowing detail to get book_id and member_id
	resp, err := http.Get(backendURL + "/borrowing?id=" + borrowingID)
	if err != nil {
		http.Error(w, "Failed to fetch borrowing detail", 500)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, string(body), resp.StatusCode)

		return
	}

	var detail BorrowingDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		http.Error(w, "Failed to decode borrowing detail", 500)
		return
	}

	payload := map[string]int{"book_id": detail.BookID, "member_id": detail.MemberID}
	b, _ := json.Marshal(payload)

	resp2, err := http.Post(backendURL+"/return", "application/json", bytes.NewReader(b))
	if err != nil {
		http.Error(w, "Failed to return book", 500)
		return
	}

	defer resp2.Body.Close()

	if resp2.StatusCode != 204 {
		body, _ := io.ReadAll(resp2.Body)
		http.Error(w, string(body), resp2.StatusCode)

		return
	}

	http.Redirect(w, r, "/borrowed", http.StatusSeeOther)
}

func borrowBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	fmt.Printf("jsonPayload: %v\n", string(jsonPayload))

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

func isbnLookupPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		isbn := r.URL.Query().Get("isbn")
		if isbn == "" {
			templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": ""})
			return
		}

		resp, err := http.Get(backendURL + "/isbn?isbn=" + isbn)

		fmt.Println("get isbn lookup page")
		fmt.Printf("resp: %+v\n", resp)
		fmt.Printf("err: %v\n", err)

		if err != nil {
			templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": isbn, "Error": "Failed to contact backend"})
			return
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": isbn, "Error": string(body)})

			return
		}

		defer resp.Body.Close()

		var pretty string

		var raw map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&raw); err == nil {
			b, _ := json.MarshalIndent(raw, "", "  ")
			pretty = string(b)
		} else {
			pretty = "Failed to decode response"
		}

		templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": isbn, "Result": pretty})

		return
	}

	if r.Method == http.MethodPost {
		isbn := r.FormValue("isbn")
		if isbn == "" {
			templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": "", "Error": "Missing ISBN"})
			return
		}
		// Fetch info from backend
		resp, err := http.Get(backendURL + "/isbn?isbn=" + isbn)

		fmt.Println("post isbn lookup page")
		fmt.Printf("resp: %+v\n", resp)
		fmt.Printf("err: %v\n", err)

		if err != nil {
			templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": isbn, "Error": "Failed to contact backend"})
			return
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": isbn, "Error": string(body)})

			return
		}

		defer resp.Body.Close()

		var book Book
		if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
			templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": isbn, "Error": "Failed to decode response"})
			return
		}
		// Prepare minimal book struct for backend
		// newBook := map[string]interface{}{
		// 	"title":            book["title"],
		// 	"author":           "",
		// 	"isbn":             isbn,
		// 	"genre":            "",
		// 	"publication_year": book["publish_date"],
		// 	"copies_total":     1,
		// 	"copies_available": 1,
		// }

		// if authors, ok := book["authors"].([]interface{}); ok && len(authors) > 0 {
		// 	if authorMap, ok := authors[0].(map[string]interface{}); ok {
		// 		newBook["author"] = authorMap["name"]
		// 	}
		// }

		b, _ := json.Marshal(book)

		resp2, err := http.Post(backendURL+"/books", "application/json", bytes.NewReader(b))
		if err != nil {
			templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": isbn, "Error": "Failed to save book"})
			return
		}

		defer resp2.Body.Close()

		if resp2.StatusCode != 200 {
			body, _ := io.ReadAll(resp2.Body)
			templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": isbn, "Error": string(body)})

			return
		}

		templates.ExecuteTemplate(w, "isbn_lookup.gohtml", map[string]interface{}{"ISBN": isbn, "Result": "Book saved to library!"})
	}
}

func deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bookID := r.FormValue("book_id")
	if bookID == "" {
		http.Error(w, "Missing book ID", 400)
		return
	}

	req, err := http.NewRequest(http.MethodDelete, backendURL+"/books?id="+bookID, nil)
	if err != nil {
		http.Error(w, "Failed to create request", 500)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to contact backend", 500)
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

func updateBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bookID := r.FormValue("book_id")
	copiesTotal := r.FormValue("copies_total")

	if bookID == "" || copiesTotal == "" {
		http.Error(w, "Missing fields", 400)
		return
	}
	// Fetch the book to get all fields
	resp, err := http.Get(backendURL + "/books")
	if err != nil {
		http.Error(w, "Failed to fetch book", 500)
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
		if strconv.Itoa(b.ID) == bookID {
			book = &b
			break
		}
	}

	if book == nil {
		http.Error(w, "Book not found", 404)
		return
	}
	// Update copies total
	var ct int

	fmt.Sscanf(copiesTotal, "%d", &ct)

	book.CopiesTotal = ct
	if book.CopiesAvailable > ct {
		book.CopiesAvailable = ct // Don't allow more available than total
	}

	b, _ := json.Marshal(book)

	req, err := http.NewRequest(http.MethodPut, backendURL+"/books", bytes.NewReader(b))
	if err != nil {
		http.Error(w, "Failed to create request", 500)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to update book", 500)
		return
	}

	defer resp2.Body.Close()

	if resp2.StatusCode != 204 {
		body, _ := io.ReadAll(resp2.Body)
		http.Error(w, string(body), resp2.StatusCode)

		return
	}

	http.Redirect(w, r, "/book?id="+bookID, http.StatusSeeOther)
}

// Handler for the reports UI page
func reportsPage(w http.ResponseWriter, r *http.Request) {
	borrowedResp, err := http.Get(backendURL + "/reports/borrowed")
	if err != nil {
		http.Error(w, "Failed to fetch borrowed books", 500)
		return
	}

	defer borrowedResp.Body.Close()

	var borrowed []map[string]interface{}
	if err := json.NewDecoder(borrowedResp.Body).Decode(&borrowed); err != nil {
		http.Error(w, "Failed to decode borrowed books", 500)
		return
	}

	overdueResp, err := http.Get(backendURL + "/reports/overdue")
	if err != nil {
		http.Error(w, "Failed to fetch overdue books", 500)
		return
	}

	defer overdueResp.Body.Close()

	var overdue []map[string]interface{}
	if err := json.NewDecoder(overdueResp.Body).Decode(&overdue); err != nil {
		http.Error(w, "Failed to decode overdue books", 500)
		return
	}

	membersResp, err := http.Get(backendURL + "/members")
	if err != nil {
		http.Error(w, "Failed to fetch members", 500)
		return
	}

	defer membersResp.Body.Close()

	var members []map[string]interface{}
	if err := json.NewDecoder(membersResp.Body).Decode(&members); err != nil {
		http.Error(w, "Failed to decode members", 500)
		return
	}

	data := map[string]interface{}{
		"Borrowed": borrowed,
		"Overdue":  overdue,
		"Members":  members,
	}

	err = templates.ExecuteTemplate(w, "reports.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

// Handler for member detail page
func memberDetailPage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing member id", 400)
		return
	}

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

	var member *Member

	for _, m := range members {
		if strconv.Itoa(m.ID) == id {
			member = &m
			break
		}
	}

	if member == nil {
		http.Error(w, "Member not found", 404)
		return
	}

	err = templates.ExecuteTemplate(w, "member_detail.gohtml", map[string]interface{}{"Member": member})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

// Handler for updating member details
func updateMemberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	name := r.FormValue("name")
	contact := r.FormValue("contact")

	memberID := r.FormValue("member_id")
	if id == "" || name == "" || memberID == "" {
		http.Error(w, "Missing fields", 400)
		return
	}

	mid, _ := strconv.Atoi(id)
	m := Member{
		ID:       mid,
		Name:     name,
		Contact:  contact,
		MemberID: memberID,
	}
	b, _ := json.Marshal(m)

	req, err := http.NewRequest(http.MethodPut, backendURL+"/members", bytes.NewReader(b))
	if err != nil {
		http.Error(w, "Failed to create request", 500)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to update member", 500)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, string(body), resp.StatusCode)

		return
	}

	http.Redirect(w, r, "/member?id="+id, http.StatusSeeOther)
}

// Handler for deleting a member
func deleteMemberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "Missing member id", 400)
		return
	}

	req, err := http.NewRequest(http.MethodDelete, backendURL+"/members?id="+id, nil)
	if err != nil {
		http.Error(w, "Failed to create request", 500)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to delete member", 500)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, string(body), resp.StatusCode)

		return
	}

	http.Redirect(w, r, "/members", http.StatusSeeOther)
}
