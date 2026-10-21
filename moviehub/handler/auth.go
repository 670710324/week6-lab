package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"moviehub/auth"
	"moviehub/mw"
	"moviehub/store"
)

type registerReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	displayName := strings.TrimSpace(req.DisplayName)
	switch {
	case !strings.Contains(email, "@") || len(email) > 255:
		writeError(w, http.StatusBadRequest, "BAD_EMAIL", "รูปแบบอีเมลไม่ถูกต้อง")
		return
	case utf8.RuneCountInString(req.Password) < 8 || len(req.Password) > 72:
		writeError(w, http.StatusBadRequest, "BAD_PASSWORD", "รหัสผ่านต้องยาว 8 ตัวอักษรขึ้นไป และไม่เกิน 72 ไบต์")
		return
	case displayName == "" || utf8.RuneCountInString(displayName) > 50:
		writeError(w, http.StatusBadRequest, "BAD_NAME", "ต้องมีชื่อที่แสดง และยาวไม่เกิน 50 ตัวอักษร")
		return
	}
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		internal(w, err)
		return
	}
	member, err := h.Store.CreateMember(r.Context(), email, passwordHash, displayName)
	if errors.Is(err, store.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "EMAIL_TAKEN", "อีเมลนี้มีบัญชีอยู่แล้ว")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, member)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	member, err := h.Store.MemberByEmail(r.Context(), email)
	if len(req.Password) > 72 {
		if h.DummyHash == "" {
			internal(w, errors.New("dummy password hash is not configured"))
			return
		}
		auth.CheckPassword(h.DummyHash, strings.Repeat("x", 72))
		writeError(w, http.StatusUnauthorized, "BAD_CREDENTIALS", "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		if h.DummyHash == "" {
			internal(w, errors.New("dummy password hash is not configured"))
			return
		}
		auth.CheckPassword(h.DummyHash, req.Password)
		writeError(w, http.StatusUnauthorized, "BAD_CREDENTIALS", "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	if !auth.CheckPassword(member.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "BAD_CREDENTIALS", "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	token, err := h.Tickets.Issue(member.ID, member.Role)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "member": member})
}

// Me (ด่านที่ 4) ตอบข้อมูลของคนที่ถือตั๋ว เลขสมาชิกมาจากตั๋วที่ RequireLogin ตรวจแล้วเท่านั้น
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	id, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}
	m, err := h.Store.MemberByID(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "MEMBER_GONE", "ไม่พบบัญชีนี้แล้ว กรุณาเข้าสู่ระบบใหม่")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}
