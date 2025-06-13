package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/library-ils-backend/internal/model"
)

type ISBN_Response struct {
	Title          string       `json:"title"`
	Authors        []Authors    `json:"authors"`
	PublishDate    string       `json:"publish_date"`
	Type           Type         `json:"type"`
	Isbn10         []string     `json:"isbn_10"`
	Isbn13         []string     `json:"isbn_13"`
	LocalID        []string     `json:"local_id"`
	Publishers     []string     `json:"publishers"`
	SourceRecords  []string     `json:"source_records"`
	Ocaid          string       `json:"ocaid"`
	Key            string       `json:"key"`
	Works          []Works      `json:"works"`
	Covers         []int        `json:"covers"`
	LatestRevision int          `json:"latest_revision"`
	Revision       int          `json:"revision"`
	Created        Created      `json:"created"`
	LastModified   LastModified `json:"last_modified"`
}
type Authors struct {
	Key string `json:"key"`
}
type Type struct {
	Key string `json:"key"`
}
type Works struct {
	Key string `json:"key"`
}
type Created struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type LastModified struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type AuthorResponse struct {
	Name         string       `json:"name"`
	PersonalName string       `json:"personal_name"`
	LastModified LastModified `json:"last_modified"`
	Key          string       `json:"key"`
	Type         Type         `json:"type"`
	ID           int          `json:"id"`
	Revision     int          `json:"revision"`
}

// isbnInfoHandler fetches book info from Open Library API by ISBN
func (s *Service) isbnInfoHandler(w http.ResponseWriter, r *http.Request) {
	isbn := chi.URLParam(r, "isbn")
	if isbn == "" {
		http.Error(w, "Missing isbn parameter", http.StatusBadRequest)
		return
	}

	book, err := s.lookupByISBN(isbn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching book info: %v", err), http.StatusBadRequest)
		return
	}

	writeJSON(w, book)
}

func (s *Service) lookupByISBN(isbn string) (*model.Book, error) {
	book, err := s.repository.SearchBookByISBN(isbn)
	if err == nil && book != nil {
		fmt.Printf("Found book in local store: %+v\n", book)
		return book, nil
	}

	isbnInfo, err := lookupBook(isbn)
	fmt.Printf("isbnInfo: %+v\n", isbnInfo)

	if err != nil {
		fmt.Printf("isbnInfo err: %v\n", err)
		return nil, fmt.Errorf("error looking up book by ISBN: %v", err)
	}

	authorName := "Unknown"

	if len(isbnInfo.Authors) > 0 {
		authorKey := isbnInfo.Authors[0].Key

		authorInfo, err := lookupAuthor(authorKey)
		if err != nil {
			return nil, fmt.Errorf("error looking up author: %v", err)
		}

		authorName = authorInfo.Name
	}

	year, err := strconv.Atoi(isbnInfo.Created.Value[:4])
	if err != nil {
		return nil, fmt.Errorf("error parsing publication (%v, [:4]: %v) year: %v", isbnInfo.Created.Value, isbnInfo.Created.Value[:4], err)
	}

	book = &model.Book{
		Title:           isbnInfo.Title,
		Author:          authorName,
		ISBN:            isbn,
		PublicationYear: year,
		CopiesTotal:     1,
		CopiesAvailable: 1,
	}

	return book, nil
}

func lookupBook(isbn string) (*ISBN_Response, error) {
	url := fmt.Sprintf("https://openlibrary.org/isbn/%s.json", isbn)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch info from Open Library: status: %v, err: %v", resp.StatusCode, err)
	}

	defer resp.Body.Close()

	var isbnResp ISBN_Response
	if err := json.NewDecoder(resp.Body).Decode(&isbnResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &isbnResp, nil
}

func lookupAuthor(authorKey string) (*AuthorResponse, error) {
	authorKey = strings.TrimPrefix(authorKey, "/authors/")
	url := fmt.Sprintf("https://openlibrary.org/authors/%s.json", authorKey)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch author info: status: %v, err: %v", resp.StatusCode, err)
	}

	defer resp.Body.Close()

	var author AuthorResponse
	if err := json.NewDecoder(resp.Body).Decode(&author); err != nil {
		return nil, fmt.Errorf("failed to decode author response: %v", err)
	}

	return &author, nil
}
