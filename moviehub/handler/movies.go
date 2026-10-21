package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"moviehub/mw"
	"moviehub/store"
)

func (h *Handler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.SearchMovies(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) GetMovie(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	movie, err := h.Store.GetMovie(r.Context(), movieID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movie)
}

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	items, err := h.Store.ListReviews(r.Context(), movieID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) AddReview(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	text := strings.TrimSpace(req.Text)
	if utf8.RuneCountInString(text) < 10 {
		writeError(w, http.StatusBadRequest, "REVIEW_TOO_SHORT", "รีวิวสั้นเกินไป")
		return
	}
	who := mw.WhoFrom(r.Context())
	review, err := h.Store.AddReview(r.Context(), who.ID, movieID, text)
	switch {
	case errors.Is(err, store.ErrDuplicate):
		writeError(w, http.StatusConflict, "ALREADY_REVIEWED", "คุณรีวิวเรื่องนี้ไปแล้ว")
		return
	case errors.Is(err, store.ErrNoMovie):
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	case err != nil:
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": review.ID, "text": review.Text, "createdAt": review.CreatedAt,
	})
}

func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	reviewID, ok := pathID(r, "reviewId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขรีวิวต้องเป็นตัวเลข")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.DeleteReview(r.Context(), reviewID, who.ID, who.Role == "admin")
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "REVIEW_NOT_FOUND", "ไม่พบรีวิวนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListWishlist(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	items, err := h.Store.ListWishlist(r.Context(), who.ID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) PutWishlistItem(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "movieId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	var req struct {
		Note string `json:"note"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.PutWishlistItem(r.Context(), who.ID, movieID, strings.TrimSpace(req.Note))
	if errors.Is(err, store.ErrNoMovie) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteWishlistItem(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "movieId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	who := mw.WhoFrom(r.Context())
	if err := h.Store.DeleteWishlistItem(r.Context(), who.ID, movieID); err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PutVote(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	var req struct {
		Score int16 `json:"score"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	if req.Score < 1 || req.Score > 10 {
		writeError(w, http.StatusBadRequest, "BAD_SCORE", "คะแนนต้องอยู่ระหว่าง 1 ถึง 10")
		return
	}
	who := mw.WhoFrom(r.Context())
	if err := h.Store.PutVote(r.Context(), who.ID, movieID, req.Score); errors.Is(err, store.ErrNoMovie) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	} else if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteVote(w http.ResponseWriter, r *http.Request) {
	movieID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}
	who := mw.WhoFrom(r.Context())
	if err := h.Store.DeleteVote(r.Context(), who.ID, movieID); err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
