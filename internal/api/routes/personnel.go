package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shiftschedule/internal/clients/postgres"
)

type PersonnelHandler struct {
	pg *postgres.Postgres
}

func (p *PersonnelHandler) GetPersonnelAll(w http.ResponseWriter, r *http.Request) error {
	personnel, err := p.pg.GetPersonnel()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(personnel)
	if err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}

func (p *PersonnelHandler) GetPersonnelByName(w http.ResponseWriter, r *http.Request) error {
	personnel, err := p.pg.GetPersonnelByName(chi.URLParam(r, "name"))
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(personnel)
	if err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}

func (p *PersonnelHandler) NewPersonnel(w http.ResponseWriter, r *http.Request) error {

	var input struct {
		Names []string `json:"names" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, "failed to bind input", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	if len(input.Names) == 0 {

	}
	for i, name := range input.Names {
		if name == "" {
			writeJSONError(w, fmt.Sprintf("input name entry indexed %d is empty", i), http.StatusBadRequest)
			return fmt.Errorf("input name entry indexed %d is empty", i)
		}
	}

	err := p.pg.NewPersonnel(input.Names)
	if err != nil {
		writeJSONError(w, "failed to update database", http.StatusInternalServerError)
		return err
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

func (p *PersonnelHandler) UpdatePersonnel(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (p *PersonnelHandler) DeletePersonnel(w http.ResponseWriter, r *http.Request) error {
	name := chi.URLParam(r, "name")
	if name != "" {
		return errors.New("mandatory parameter 'name' is empty or missing")
	}
	err := p.pg.DeletePersonnel(name)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	return nil
}

// import "github.com/go-playground/validator/v10"
//
// var validate = validator.New()
//
// type CreateUserRequest struct {
//     Name  string `json:"name" validate:"required"`
//     Email string `json:"email" validate:"required,email"`
// }
//
// func CreateUser(w http.ResponseWriter, r *http.Request) {
//     var req CreateUserRequest
//     if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//         http.Error(w, "invalid JSON", http.StatusBadRequest)
//         return
//     }
//
//     if err := validate.Struct(req); err != nil {
//         http.Error(w, err.Error(), http.StatusBadRequest)
//         return
//     }
//
//     // Continue…
// }

// pageStr := r.URL.Query().Get("page")
// if pageStr == "" {
//     http.Error(w, "missing page parameter", http.StatusBadRequest)
//     return
// }
