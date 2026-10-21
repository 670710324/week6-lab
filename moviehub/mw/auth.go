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

// MemberID คืนเลขสมาชิกจากตั๋วที่ตรวจแล้ว ok = false แปลว่าคำขอนี้ไม่ได้ผ่าน RequireLogin
func MemberID(ctx context.Context) (int64, bool) {
	who := WhoFrom(ctx)
	return who.ID, who.ID > 0
}

// Role คืนบทบาทจากตั๋ว ("member" หรือ "admin") ว่างถ้ายังไม่ได้ login
func Role(ctx context.Context) string {
	return WhoFrom(ctx).Role
}

func fail(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

// RequireLogin คือพนักงานตรวจตั๋ว: ต้องมี Authorization: Bearer <token> ที่ตรวจผ่าน
func RequireLogin(iss *auth.Issuer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := strings.Fields(r.Header.Get("Authorization"))
		if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
			fail(w, http.StatusUnauthorized, "LOGIN_REQUIRED", "กรุณาเข้าสู่ระบบก่อน")
			return
		}
		claims, err := iss.Parse(header[1])
		if err != nil {
			fail(w, http.StatusUnauthorized, "TOKEN_INVALID", "ตั๋วใช้ไม่ได้หรือหมดอายุ กรุณาเข้าสู่ระบบใหม่")
			return
		}
		memberID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil || memberID <= 0 {
			fail(w, http.StatusUnauthorized, "TOKEN_INVALID", "ตั๋วใช้ไม่ได้หรือหมดอายุ กรุณาเข้าสู่ระบบใหม่")
			return
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, Who{ID: memberID, Role: claims.Role})
		next(w, r.WithContext(ctx))
	}
}

// RequireRole ใช้ต่อจาก RequireLogin: ตั๋วถูกต้องแต่บทบาทไม่ตรง ตอบ 403
func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if Role(r.Context()) != role {
			fail(w, http.StatusForbidden, "FORBIDDEN", "ต้องเป็น "+role+" เท่านั้น")
			return
		}
		next(w, r)
	}
}
