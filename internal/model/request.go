package model

type BorrowRequest struct {
	BookID   string `json:"book_id"`
	MemberID string `json:"member_id"`
}
