package frontend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tliefheid/go-ils/internal/model"
)

func (s *Service) memberPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("memberPost called")

	idStr := r.FormValue("id")
	name := r.FormValue("name")
	contact := r.FormValue("contact")

	if name == "" || idStr == "" || contact == "" {
		http.Error(w, "Missing fields", 400)
		return
	}

	id := 0

	fmt.Printf("idStr: %v\n", idStr)
	fmt.Printf("name: %v\n", name)
	fmt.Printf("contact: %v\n", contact)

	if idStr != "new" {
		var err error

		id, err = strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid member ID", http.StatusBadRequest)
			return
		}
		// Create new member
	}

	member := model.Member{
		ID:      id,
		Name:    name,
		Contact: contact,
	}

	m, _ := json.Marshal(member)

	if idStr == "new" {
		fmt.Println("Creating new member")
		// New member, send POST request to create
		err := s.newMember(m)
		if err != nil {
			s.errorPage(w, "Failed to create new member", err)
			return
		}
	} else {
		fmt.Println("Updating existing member")
		// Existing member, send PUT request to update
		err := s.updateMember(idStr, m)
		if err != nil {
			s.errorPage(w, "Failed to update member", err)
			return
		}
	}

	s.memberPage(w, r)
}

func (s *Service) newMember(m []byte) error {
	resp, err := http.Post(s.uri+"/members", "application/json", bytes.NewReader(m))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return errors.New("Invalid response status code: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func (s *Service) updateMember(id string, m []byte) error {
	client := &http.Client{}

	req, err := http.NewRequest("PUT", s.uri+"/members/"+id, bytes.NewReader(m))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Invalid response status code: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func (s *Service) memberPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var resp *http.Response

	var err error
	if q != "" {
		resp, err = http.Get(s.uri + "/members/search?q=" + q)
	} else {
		resp, err = http.Get(s.uri + "/members")
	}

	if err != nil {
		s.errorPage(w, "Failed to fetch members", err)
		return
	}

	defer resp.Body.Close()

	var members []model.Member

	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		s.errorPage(w, "Failed to decode members", err)
		return
	}

	s.executeTemplate(w, "members.gohtml", map[string]interface{}{
		"Members": members,
		"Query":   q,
	})
}

type memberDetailData struct {
	IsNew  bool
	Member model.Member
}

func (s *Service) memberDetailPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		s.errorPage(w, "Missing member id", nil)
		return
	}

	if id == "new" {
		// New member
		s.executeTemplate(w, "member_upsert.gohtml", memberDetailData{
			IsNew:  true,
			Member: model.Member{},
		})

		return
	}

	resp, err := http.Get(s.uri + "/members/" + id)
	if err != nil {
		s.errorPage(w, "Failed to fetch member", err)
		return
	}

	defer resp.Body.Close()

	var member model.Member

	if err := json.NewDecoder(resp.Body).Decode(&member); err != nil {
		s.errorPage(w, "Failed to decode member", err)
		return
	}

	s.executeTemplate(w, "member_upsert.gohtml", memberDetailData{
		IsNew:  false,
		Member: member,
	})
}

func (s *Service) memberDeletePost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		s.errorPage(w, "Missing member id", nil)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("DELETE", s.uri+"/members/"+id, nil)
	if err != nil {
		s.errorPage(w, "Failed to create delete request", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		s.errorPage(w, "Failed to delete member", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		s.errorPage(w, "Failed to delete member", fmt.Errorf("invalid response status code: %d", resp.StatusCode))
		return
	}

	s.memberPage(w, r)
}
