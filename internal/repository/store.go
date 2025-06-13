package repository

import (
	"errors"

	"github.com/tliefheid/go-ils/internal/model"
)

var (
	ErrNotFound = errors.New("not found")
)

type Store interface {
	BookStore
	MemberStore
	BorrowingStore

	Migrate(fn string) error
	Close() error
}

type BookStore interface {
	ListBooks() ([]model.Book, error)
	SearchBookByISBN(isbn string) (*model.Book, error)
	SearchBooks(search string) ([]model.Book, error)
	AddBook(book model.Book) error
	GetBook(id int) (*model.Book, error)
	UpdateBook(book model.Book) error
	DeleteBook(id int) error
}
type MemberStore interface {
	ListMemberss() ([]model.Member, error)
	SearchMembers(search string) ([]model.Member, error)
	AddMember(member model.Member) error
	GetMember(id int) (*model.Member, error)
	UpdateMember(member model.Member) error
	DeleteMember(id int) error
	// ListMembers lists all members in the store.
}

type BorrowingStore interface {
	ListBorrowings() ([]model.BorrowingDetail, error)
	AddBorrowing(borrowing model.Borrowing) error
	GetBorrowing(id int) (*model.BorrowingDetail, error)
	ReturnBorrowing(id int) error
	// UpdateBorrowing(borrowing model.Borrowing) error
	DeleteBorrowing(id int) error
}
