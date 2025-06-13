package postgres

import (
	"fmt"

	"github.com/tliefheid/go-ils/internal/model"
)

func (s *Store) ListMemberss() ([]model.Member, error) {
	rows, err := s.db.Query("SELECT id, name, contact FROM members")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var members []model.Member

	for rows.Next() {
		var m model.Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Contact); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		members = append(members, m)
	}

	return members, nil
}

func (s *Store) SearchMembers(search string) ([]model.Member, error) {
	rows, err := s.db.Query(`SELECT id, name, contact FROM members WHERE name ILIKE '%' || $1 || '%' OR contact ILIKE '%' || $1 || '%'`, search)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var members []model.Member

	for rows.Next() {
		var m model.Member

		if err := rows.Scan(&m.ID, &m.Name, &m.Contact); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		members = append(members, m)
	}

	return members, nil
}

func (s *Store) AddMember(member model.Member) error {
	query := `INSERT INTO members (name, contact) VALUES ($1, $2)`

	_, err := s.db.Query(query, member.Name, member.Contact)
	if err != nil {
		return err
	}

	return nil
}
func (s *Store) GetMember(id int) (*model.Member, error) {
	rows, err := s.db.Query("SELECT id, name, contact FROM members WHERE id=$1", id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var members []*model.Member

	for rows.Next() {
		var m model.Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Contact); err != nil {
			continue
		}

		members = append(members, &m)
	}

	return members[0], nil
}
func (s *Store) UpdateMember(m model.Member) error {
	query := `UPDATE members SET name=$1, contact=$2,WHERE id=$3`

	_, err := s.db.Exec(query, m.Name, m.Contact, m.ID)
	if err != nil {
		return err
	}

	return nil
}
func (s *Store) DeleteMember(id int) error {
	// First, delete all borrowings for this member (to avoid FK constraint errors)
	_, err := s.db.Exec("DELETE FROM borrowings WHERE member_id=$1", id)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM members WHERE id=$1", id)
	if err != nil {
		return err
	}

	return nil
}
