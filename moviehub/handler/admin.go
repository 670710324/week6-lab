package handler

import (
	"errors"
	"io"
	"net/http"

	"moviehub/mw"
	"moviehub/store"
)

func (h *Handler) AdminMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.Store.ListMembers(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": members})
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
	if err := decodeJSON(w, r, &req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	if len(req.Note) > 500 {
		writeError(w, http.StatusBadRequest, "BAD_NOTE", "หมายเหตุต้องไม่เกิน 500 ไบต์")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.PutWishlistItem(r.Context(), who.ID, movieID, req.Note)
	if errors.Is(err, store.ErrNoMovie) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"movieId": movieID, "note": req.Note})
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
	err := h.Store.PutVote(r.Context(), who.ID, movieID, req.Score)
	if errors.Is(err, store.ErrNoMovie) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"movieId": movieID, "score": req.Score})
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
