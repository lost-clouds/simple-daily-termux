package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"simple-daily-termux/internal/calendar"
	"simple-daily-termux/internal/config"
	"simple-daily-termux/internal/countdown"
	"simple-daily-termux/internal/diary"
	"simple-daily-termux/internal/ledger"
	"simple-daily-termux/internal/pomodoro"
	"simple-daily-termux/internal/store/sqlstore"
	"simple-daily-termux/internal/summary"
	"simple-daily-termux/internal/todo"
)

//go:embed all:web
var webFS embed.FS

func main() {
	cfgPath := "./config.json"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Initialize store
	var st sqlstore.Store
	switch cfg.Database.Driver {
	case "sqlite":
		st, err = sqlstore.NewSQLite(cfg.Database.SQLite.Path)
	case "mysql":
		st, err = sqlstore.NewMySQL(cfg.Database.MySQL.DSN)
	default:
		log.Fatalf("unsupported database driver: %s", cfg.Database.Driver)
	}
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()

	// Wire up services
	ledgerSvc := ledger.NewService(st.Ledgers(), st.Settings())
	countSvc := countdown.NewService(st.Countdowns())
	todoSvc := todo.NewService(st.Todos(), countSvc)
	pomoSvc := pomodoro.NewService(st.Pomodoros())
	diarySvc := diary.NewService(st.Diaries(), ledgerSvc)
	calSvc := calendar.NewService(st.Calendars(), st.Todos(), st.Countdowns(), st.Diaries())
	summarySvc := summary.NewService(ledgerSvc, countSvc, pomoSvc, cfg.Database.Timezone)

	// Routes
	mux := http.NewServeMux()

	todo.RegisterHandler(mux, todoSvc)
	countdown.RegisterHandler(mux, countSvc)
	pomodoro.RegisterHandler(mux, pomoSvc, cfg.Database.Timezone)
	diary.RegisterHandler(mux, diarySvc)
	ledger.RegisterHandler(mux, ledgerSvc)
	ledger.RegisterSettingsHandler(mux, ledgerSvc)
	calendar.RegisterHandler(mux, calSvc)
	summary.RegisterHandler(mux, summarySvc)

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ok":true,"data":{"status":"healthy"}}`)
	})

	// /simpledaily/ redirects to main SPA (for Blog-termux card link)
	mux.HandleFunc("/simpledaily/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})

	// Static files
	webSub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("embed: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webSub)))

	// Middleware
	var handler http.Handler = mux
	handler = logging(handler)
	if cfg.Server.CORS {
		handler = cors(handler)
	}

	// Server
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("simple-daily-termux listening on %s", cfg.Server.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-stop
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("stopped")
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
