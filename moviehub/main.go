package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"moviehub/auth"
	"moviehub/handler"
	"moviehub/mw"
	"moviehub/store"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("ต้องกำหนด DATABASE_URL")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("ต่อฐานข้อมูลไม่ติด ", err)
	}

	dummyHash, err := auth.HashPassword("not-a-real-moviehub-password")
	if err != nil {
		log.Fatal("สร้างลายนิ้วมือสำหรับตรวจการเข้าสู่ระบบไม่สำเร็จ ", err)
	}
	tickets := &auth.Issuer{Secret: []byte(secret), TTL: time.Hour}
	h := &handler.Handler{
		Store:     &store.Store{Pool: pool},
		Tickets:   tickets,
		DummyHash: dummyHash,
	}
	login := func(f http.HandlerFunc) http.HandlerFunc {
		return mw.RequireLogin(tickets, f)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("GET /api/movies", h.SearchMovies)
	mux.HandleFunc("GET /api/movies/{id}/reviews", h.ListReviews)
	mux.HandleFunc("GET /api/me", login(h.Me))
	mux.HandleFunc("POST /api/movies/{id}/reviews", login(h.AddReview))
	mux.HandleFunc("DELETE /api/reviews/{reviewId}", login(h.DeleteReview))
	mux.HandleFunc("GET /api/admin/members", login(mw.RequireRole("admin", h.AdminMembers)))
	mux.HandleFunc("GET /api/me/wishlist", login(h.ListWishlist))
	mux.HandleFunc("PUT /api/me/wishlist/{movieId}", login(h.PutWishlistItem))
	mux.HandleFunc("DELETE /api/me/wishlist/{movieId}", login(h.DeleteWishlistItem))
	mux.HandleFunc("PUT /api/movies/{id}/vote", login(h.PutVote))
	mux.HandleFunc("DELETE /api/movies/{id}/vote", login(h.DeleteVote))

	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	server := &http.Server{
		Addr:              ":8080",
		Handler:           mw.Logging(mw.CORS(allowedOrigins, mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()
	log.Println("MovieHub API ฟังอยู่ที่ :8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
		return
	case <-quit:
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("ปิดเซิร์ฟเวอร์ไม่ทัน: %v", err)
	}
}
