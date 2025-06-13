package frontend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tliefheid/go-ils/internal/model"
)

var (
	isbnRegex = regexp.MustCompile(`^[\d-]{10,17}$`)
)

func (s *Service) bookPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("bookPost called")

	idStr := r.FormValue("id")
	title := r.FormValue("title")
	contact := r.FormValue("author")
	isbn := r.FormValue("isbn")
	pubYear := r.FormValue("publication_year")
	copies := r.FormValue("copies_total")

	fmt.Printf("idStr: %v\n", idStr)
	fmt.Printf("title: %v\n", title)
	fmt.Printf("contact: %v\n", contact)
	fmt.Printf("isbn: %v\n", isbn)
	fmt.Printf("pubYear: %v\n", pubYear)
	fmt.Printf("copies: %v\n", copies)

	if idStr == "" ||
		title == "" ||
		contact == "" ||
		isbn == "" ||
		pubYear == "" ||
		copies == "" || !isbnRegex.MatchString(isbn) {
		r.Method = http.MethodGet
		id := 0

		if idStr != "new" {
			var err error

			id, err = strconv.Atoi(idStr)
			if err != nil {
				s.errorPage(w, "Invalid book ID", err)
				return
			}
		}

		pubYearInt, err := strconv.Atoi(pubYear)
		if err != nil {
			s.errorPage(w, "Invalid publication year", err)
			return
		}

		copiesInt, err := strconv.Atoi(copies)
		if err != nil {
			s.errorPage(w, "Invalid copies value", err)
			return
		}

		book := model.Book{
			ID:              id,
			Title:           title,
			Author:          contact,
			ISBN:            isbn,
			PublicationYear: pubYearInt,
			CopiesTotal:     copiesInt,
			CopiesAvailable: copiesInt, // Initially all copies are available
		}

		fmt.Printf("r.URL.String(): %v\n", r.URL.String())

		r.URL.RawQuery = ""
		fmt.Printf("r.URL after: %v\n", r.URL.String())

		ctx := context.WithValue(r.Context(), "book", book)
		fmt.Printf("ctx: %v\n", ctx)
		s.bookUpsertPage(w, r.WithContext(ctx))

		return
	}

	id := 0

	if idStr != "new" {
		var err error

		id, err = strconv.Atoi(idStr)
		if err != nil {
			s.errorPage(w, "Invalid book ID", err)
			return
		}
	}

	pubYearInt, err := strconv.Atoi(pubYear)
	if err != nil {
		s.errorPage(w, "Invalid publication year", err)
		return
	}

	copiesInt, err := strconv.Atoi(copies)
	if err != nil {
		s.errorPage(w, "Invalid copies value", err)
		return
	}

	book := model.Book{
		ID:              id,
		Title:           title,
		Author:          contact,
		ISBN:            isbn,
		PublicationYear: pubYearInt,
		CopiesTotal:     copiesInt,
		CopiesAvailable: copiesInt, // Initially all copies are available
	}

	b, _ := json.Marshal(book)

	if idStr == "new" {
		fmt.Println("Creating new book")

		fmt.Printf("Book to create: %+v\n", book)

		// New member, send POST request to create
		err := s.newBook(b)
		if err != nil {
			s.errorPage(w, "Failed to create new book", err)
			return
		}
	} else {
		fmt.Println("Updating existing book")
		fmt.Printf("Book to update: %+v\n", book)

		// Existing member, send PUT request to update
		err := s.updateBook(idStr, b)
		if err != nil {
			s.errorPage(w, "Failed to update member", err)
			return
		}
	}

	s.booksPage(w, r)
}

func (s *Service) newBook(b []byte) error {
	resp, err := http.Post(s.uri+"/books", "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Printf("post book creation err: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes := new(bytes.Buffer)
		_, readErr := bodyBytes.ReadFrom(resp.Body)

		if readErr == nil {
			fmt.Printf("Response body: %s\n", bodyBytes.String())
		}

		return errors.New("Invalid response status code: " + strconv.Itoa(resp.StatusCode) + ", because: " + bodyBytes.String())
	}

	return nil
}

func (s *Service) updateBook(id string, b []byte) error {
	client := &http.Client{}

	req, err := http.NewRequest("PUT", s.uri+"/books/"+id, bytes.NewReader(b))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bodyBytes := new(bytes.Buffer)
	_, readErr := bodyBytes.ReadFrom(resp.Body)

	if readErr == nil {
		fmt.Printf("Response body: %s\n", bodyBytes.String())
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.New("Invalid response status code: " + strconv.Itoa(resp.StatusCode) + ", because: " + bodyBytes.String())
	}

	return nil
}

// should return only a boolean
func (s *Service) checkBook(isbn string) (bool, error) {
	resp, err := http.Get(s.uri + "/books/isbn/" + isbn)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil // Book not found
	}

	if resp.StatusCode == http.StatusOK {
		return true, nil // Book exists
	}
	// If we reach here, it means there was an unexpected status code

	bodyBytes := new(bytes.Buffer)
	_, err = bodyBytes.ReadFrom(resp.Body)

	if err != nil {
		bodyBytes.WriteString("Failed to read response body")
	}

	return false, errors.New("book not found, %v" + bodyBytes.String())
}

func (s *Service) booksPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var resp *http.Response

	var err error
	if q != "" {
		resp, err = http.Get(s.uri + "/books/search?q=" + q)
	} else {
		resp, err = http.Get(s.uri + "/books")
	}

	if err != nil {
		s.errorPage(w, "failed to fetch books", err)
		return
	}

	defer resp.Body.Close()

	var books []model.Book
	if err := json.NewDecoder(resp.Body).Decode(&books); err != nil {
		s.errorPage(w, "failed to decose books", err)

		return
	}

	if len(books) == 1 {
		// If only one book is found, redirect to its detail page
		http.Redirect(w, r, "/books/"+strconv.Itoa(books[0].ID), http.StatusSeeOther)
		return
	}

	s.executeTemplate(w, "books.gohtml", map[string]interface{}{
		"Books": books,
		"Query": q,
	})
}

type bookDetailData struct {
	IsNew           bool
	Book            model.Book
	ValidationError map[string]string
}

func (s *Service) bookUpsertPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("bookUpsertPage called")
	fmt.Printf("r.URL.String(): %v\n", r.URL.String())

	ctx := r.Context()
	fmt.Printf("ctx: %v\n", ctx)

	bookVal := ctx.Value("book")
	if bookVal != nil {
		book, ok := ctx.Value("book").(model.Book)
		if !ok {
			fmt.Println("bookVal is not of type model.Book")
			// break out fo this if
		} else {
			fmt.Println("bookVal is of type model.Book")
			fmt.Printf("book: %+v\n", book)
			isbnOk := isbnRegex.MatchString(book.ISBN)
			fmt.Printf("isbnOk: %v\n", isbnOk)

			validations := map[string]string{}

			if !isbnOk {
				validations["isbn"] = "Invalid ISBN format. It should be 10 to 17 characters long and can include dashes."
			}

			payload := bookDetailData{
				IsNew:           true,
				Book:            book,
				ValidationError: validations,
			}

			// check if book is already present in backend
			found, _ := s.checkBook(book.ISBN)
			if found {
				// book already exists in backend, so we set IsNew to false
				payload.IsNew = false
				payload.ValidationError = nil // Clear validation errors if book already exists
				payload.Book.CopiesAvailable++
				payload.Book.CopiesTotal++
			}

			if ok {
				// If book is provided in context, use it
				s.executeTemplate(w, "book_upsert.gohtml", payload)

				return
			}
		}
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		s.errorPage(w, "Missing book id", nil)

		return
	}

	if id == "new" {
		// New member
		s.executeTemplate(w, "book_upsert.gohtml", bookDetailData{
			IsNew: true,
			Book:  model.Book{},
		})

		return
	}

	resp, err := http.Get(s.uri + "/books/" + id)
	if err != nil {
		s.errorPage(w, "Failed to fetch book", err)
		return
	}

	defer resp.Body.Close()

	var b model.Book

	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		s.errorPage(w, "Failed to decode book", err)
		return
	}

	s.executeTemplate(w, "book_upsert.gohtml", bookDetailData{
		IsNew: false,
		Book:  b,
	})
}
func (s *Service) bookDetailPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		s.errorPage(w, "Missing book id", nil)

		return
	}

	resp, err := http.Get(s.uri + "/books/" + id)
	if err != nil {
		s.errorPage(w, "failed to fetch books", err)

		return
	}

	defer resp.Body.Close()

	var book model.Book
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		s.errorPage(w, "Failed to fetch books", err)

		return
	}

	// Fetch members for borrow dropdown
	resp2, err := http.Get(s.uri + "/members")
	if err != nil {
		s.errorPage(w, "Failed to fetch members", err)

		return
	}

	defer resp2.Body.Close()

	var members []model.Member
	if err := json.NewDecoder(resp2.Body).Decode(&members); err != nil {
		s.errorPage(w, "Failed to decode members", err)

		return
	}

	s.executeTemplate(w, "book_detail.gohtml", struct {
		Book    *model.Book
		Members []model.Member
	}{&book, members})
}

func (s *Service) deleteBookPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		s.errorPage(w, "Missing book ID", nil)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("DELETE", s.uri+"/books/"+id, nil)
	if err != nil {
		s.errorPage(w, "Failed to create delete request", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		s.errorPage(w, "Failed to delete book", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		s.errorPage(w, "Failed to delete book: "+resp.Status, nil)
		return
	}

	s.booksPage(w, r)
}
