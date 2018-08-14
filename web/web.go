package web

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/karmakaze/quicklog/config"
	"github.com/karmakaze/quicklog/storage"
	"github.com/kuangchanglang/graceful"
)

type WebServer struct {
	baseHandler http.Handler
}

func NewWebServer(baseHandler http.Handler) *WebServer {
	return &WebServer{
		baseHandler: baseHandler,
	}
}

func (ws *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws.baseHandler.ServeHTTP(w, r)
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func Serve(cfg config.Config) error {
	useSSL := cfg.SslFullChain != "" && cfg.SslPrivKey != ""

	var db *sql.DB
	if useSSL || graceful.IsWorker() {
		var err error
		db, err = storage.OpenDB(cfg)
		if err != nil {
			return err
		}
	}

	// these get added to http.DefaultServeMux
	http.Handle("/projects", NewProjectsHandler(db))
	http.Handle("/entries", NewEntriesHandler(db))
	http.HandleFunc("/status", status)

	err := runServer(cfg)
	if err != nil {
		if db != nil {
			db.Close()
		}
		log.Fatal(err.Error())
	}
	return nil
}

func runServer(cfg config.Config) error {
	if cfg.SslFullChain != "" && cfg.SslPrivKey != "" {
		// we currently don't support SSL *and* graceful restarts (use a load-balancer to get both)
		log.Printf("Listening on %v:%d (with SSL)\n", cfg.Address, cfg.Port)

		err := http.ListenAndServeTLS(cfg.Address+":"+strconv.Itoa(cfg.Port), cfg.SslFullChain, cfg.SslPrivKey, nil)
		if err != nil {
			return fmt.Errorf("ListenAndServeTLS: %v", err)
		}
	} else {
		server := graceful.NewServer()
		server.Register(cfg.Address+":"+strconv.Itoa(cfg.Port), http.DefaultServeMux)

		if graceful.IsWorker() {
			log.Printf("Listening on %v:%d (no-SSL)\n", cfg.Address, cfg.Port)
		}
		if err := server.Run(); err != nil {
			return fmt.Errorf("graceful.Server.Run: %v", err)
		}
	}
	return nil
}
