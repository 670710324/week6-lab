package mw

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"moviehub/auth"
)

type Who struct {
	ID   int64
	Role string
}

type ctxKey struct{}

func WhoFrom(ctx context.Context) Who {
	who, _ := ctx.Value(ctxKey{}).(Who)
	return who
}

func deny(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

func RequireLogin(tickets *auth.Issuer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := strings.Fields(r.Header.Get("Authorization"))
		if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
			deny(w, http.StatusUnauthorized, "LOGIN_REQUIRED", "กรุณาเข้าสู่ระบบก่อน")
			return
		}
		claims, err := tickets.Parse(header[1])
		if err != nil {
			deny(w, http.StatusUnauthorized, "TOKEN_INVALID", "ตั๋วใช้ไม่ได้หรือหมดอายุ กรุณาเข้าสู่ระบบใหม่")
			return
		}
		memberID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil || memberID <= 0 {
			deny(w, http.StatusUnauthorized, "TOKEN_INVALID", "ตั๋วใช้ไม่ได้หรือหมดอายุ กรุณาเข้าสู่ระบบใหม่")
			return
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, Who{ID: memberID, Role: claims.Role})
		next(w, r.WithContext(ctx))
	}
}

func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if WhoFrom(r.Context()).Role != role {
			deny(w, http.StatusForbidden, "FORBIDDEN", "บัญชีนี้ไม่มีสิทธิ์ทำรายการนี้")
			return
		}
		next(w, r)
	}
}
