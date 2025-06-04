package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

func LookupByISBN(isbn string) (*Book, error) {
	isbnInfo, err := lookupBook(isbn)
	if err != nil {
		return nil, fmt.Errorf("error looking up book by ISBN: %v", err)
	}

	authorKey := isbnInfo.Authors[0].Key

	authorInfo, err := lookupAuthor(authorKey)
	if err != nil {
		return nil, fmt.Errorf("error looking up author: %v", err)
	}

	year, err := strconv.Atoi(isbnInfo.Created.Value[:4])
	if err != nil {
		return nil, fmt.Errorf("error parsing publication (%v, [:4]: %v) year: %v", isbnInfo.Created.Value, isbnInfo.Created.Value[:4], err)
	}

	book := &Book{
		Title:           isbnInfo.Title,
		Author:          authorInfo.Name,
		ISBN:            isbn,
		Genre:           "Unknown", // Genre not available in Open Library API
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
