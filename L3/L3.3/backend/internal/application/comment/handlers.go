package comment

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"wb_school/L3/L3.3/backend/internal"
	"wb_school/L3/L3.3/backend/internal/service"
)

type CommentHandler struct {
	service *service.CommentService
	config  *internal.Config
	logger  *log.Logger
}

type CreateCommentRequest struct {
	Content  string  `json:"content"`
	Author   string  `json:"author"`
	ParentID *string `json:"parent_id,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewCommentHandler(service *service.CommentService, config *internal.Config, logger *log.Logger) *CommentHandler {
	return &CommentHandler{
		service: service,
		config:  config,
		logger:  logger,
	}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	comment, err := h.service.Create(r.Context(), req.Content, req.Author, req.ParentID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	parentID := r.URL.Query().Get("parent_id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	sortBy := r.URL.Query().Get("sort_by")
	order := r.URL.Query().Get("order")

	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}

	var parentIDPtr *string
	if parentID != "" {
		parentIDPtr = &parentID
	}

	tree, total, err := h.service.GetTree(r.Context(), parentIDPtr, page, pageSize, sortBy, order)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"data":  tree,
		"total": total,
		"page":  page,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.Delete(r.Context(), id); err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Comment deleted successfully"})
}

func (h *CommentHandler) SearchComments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	comments, total, err := h.service.Search(r.Context(), query, page, pageSize)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := map[string]interface{}{
		"data":  comments,
		"total": total,
		"page":  page,
		"query": query,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, ErrorResponse{Error: message})
}
