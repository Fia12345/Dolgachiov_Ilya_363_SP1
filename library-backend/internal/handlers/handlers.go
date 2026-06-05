package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"library-backend/internal/models"
	"library-backend/internal/repository"
)

type Handlers struct {
	repo     *repository.SQLiteRepository
	jwtSecret string
}

func NewHandlers(repo *repository.SQLiteRepository, jwtSecret string) *Handlers {
	return &Handlers{repo: repo, jwtSecret: jwtSecret}
}

func (h *Handlers) CreateBook(w http.ResponseWriter, r *http.Request) {
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	book.ID = uuid.New().String()
	book.Status = "Available"

	if err := h.repo.CreateBook(&book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *Handlers) GetBooks(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	author := r.URL.Query().Get("author")
	status := r.URL.Query().Get("status")

	books, err := h.repo.GetBooks(page, limit, author, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(books)
}

func (h *Handlers) GetBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	book, err := h.repo.GetBook(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(book)
}

func (h *Handlers) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	book.ID = id

	if err := h.repo.UpdateBook(&book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(book)
}

func (h *Handlers) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u.ID = uuid.New().String()
	u.RegistrationDate = time.Now()

	password := "password123"

	if err := h.repo.CreateUser(&u, password, "user"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func (h *Handlers) GetUserBooks(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	books, err := h.repo.GetUserBooks(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(books)
}

func (h *Handlers) IssueBook(w http.ResponseWriter, r *http.Request) {
	type req struct {
		BookID string `json:"book_id"`
		UserID string `json:"user_id"`
	}
	var reqData req
	json.NewDecoder(r.Body).Decode(&reqData)

	if err := h.repo.IssueBook(reqData.BookID, reqData.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "issued"})
}

func (h *Handlers) ReturnBook(w http.ResponseWriter, r *http.Request) {
	type req struct {
		BookID string `json:"book_id"`
	}
	var reqData req
	json.NewDecoder(r.Body).Decode(&reqData)

	if err := h.repo.ReturnBook(reqData.BookID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "returned"})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	json.NewDecoder(r.Body).Decode(&req)

	user, pass, role, err := h.repo.GetUserByEmail(req.Email)
	if err != nil || pass != req.Password {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token := "fake-jwt-token-for-" + user.ID + "-" + role 
	json.NewEncoder(w).Encode(map[string]string{"token": token, "role": role})
}
