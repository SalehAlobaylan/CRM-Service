package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Customer struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Contacted bool   `json:"contacted"`
}

// groups customer logic methods.
type Service struct{}

// var customers = map[int]Customer{} // i tried to use map but
// it was not the best way to do it so i used slice
var customers []Customer

func init() {
	customers = append(customers, readFromFile("DB/Customers.json")...)
}

func (s *Service) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newCustomer Customer
	json.NewDecoder(r.Body).Decode(&newCustomer) // Convert Json recieved from user to Go struct
	newCustomer.ID = int(uuid.New().ID())
	customers = append(customers, newCustomer)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newCustomer)
}

func (s *Service) GetCustomers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(customers); err != nil {
		http.Error(w, "Failed to encode customers", http.StatusInternalServerError)
		return
	}
}

func (s *Service) GetCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	for _, customer := range customers {
		if customer.ID == id {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(customer); err != nil {
				http.Error(w, "Failed to encode customer", http.StatusInternalServerError)
				return
			}
			return
		}
	}
	http.Error(w, "Customer not found", http.StatusNotFound)
}

func (s *Service) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	var updatedCustomer Customer
	if err := json.NewDecoder(r.Body).Decode(&updatedCustomer); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	for i := range customers {
		if customers[i].ID == id {
			updatedCustomer.ID = id
			customers[i] = updatedCustomer
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(updatedCustomer)
			return
		}
	}

	http.Error(w, "Customer not found", http.StatusNotFound)
}

func (s *Service) PatchCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	var payload struct { // Expect JSON body like {"contacted": true}
		Contacted bool `json:"contacted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	for i := range customers {
		if customers[i].ID == id {
			customers[i].Contacted = payload.Contacted
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(customers[i])
			return
		}
	}

	http.Error(w, "Customer not found", http.StatusNotFound)
}

func (s *Service) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	for i, customer := range customers {
		if customer.ID == id {
			customers = append(customers[:i], customers[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Customer not found", http.StatusNotFound)
}

func (s *Service) ShowContactPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.HTML")
}

// readFromFile loads customers from a JSON file; returns empty slice on error.
func readFromFile(path string) []Customer {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var list []Customer
	if err := json.Unmarshal(bytes, &list); err != nil {
		return nil
	}
	return list
}
