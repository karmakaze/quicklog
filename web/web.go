package web

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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

func findCertAndPrivKey() (string, string) {
	for _, hostname := range []string{"api.statuspage.me", "statuspage.me"} {
		certDir := "/etc/letsencrypt/live/" + hostname
		log.Printf("Looking for fullchain.pem and privkey.pem in %s\n", certDir)

		certFile := certDir + "/fullchain.pem"
		privkeyFile := certDir + "/privkey.pem"

		if _, err := os.Stat(certFile); os.IsExist(err) {
			if _, err = os.Stat(privkeyFile); os.IsExist(err) {
				log.Printf("Found fullchain.pem and privkey.pem in %s\n", certDir)
				return certFile, privkeyFile
			}
		}
	}
	return "", ""
}

var addr = flag.String("addr", "localhost:8124", "http service address")

func Serve(addr string, cfg config.Config) error {
	certFile, privkeyFile := findCertAndPrivKey()
	useSSL := certFile != "" && privkeyFile != ""

	var db *sql.DB
	if useSSL || graceful.IsWorker() {
		var err error
		db, err = storage.OpenDB(cfg)
		if err != nil {
			return err
		}
	}

	// these get added to http.DefaultServeMux
	http.HandleFunc("/status", status)
	http.Handle("/entries", NewEntriesHandler(db))

	var err error
	if useSSL {
		// we currently don't support SSL *and* graceful restarts (use a load-balancer to get both)
		err = runServer(addr, 443, certFile, privkeyFile)
	} else {
		err = runServer(addr, 8124, "", "")
	}
	if err != nil {
		if db != nil {
			db.Close()
		}
		log.Fatal(err.Error())
	}
	return nil
}

func runServer(addr string, port int, certFile, privkeyFile string) error {
	if certFile != "" && privkeyFile != "" {
		// we currently don't support SSL *and* graceful restarts (use a load-balancer to get both)
		log.Printf("Listening on %v:%d (with SSL)\n", addr, port)

		err := http.ListenAndServeTLS(addr+":"+strconv.Itoa(port), certFile, privkeyFile, nil)
		if err != nil {
			return fmt.Errorf("ListenAndServeTLS: %v", err)
		}
	} else {
		server := graceful.NewServer()
		server.Register(addr+":"+strconv.Itoa(port), http.DefaultServeMux)

		if graceful.IsWorker() {
			log.Printf("Listening on %v:%d (no-SSL)\n", addr, port)
		}
		if err := server.Run(); err != nil {
			return fmt.Errorf("graceful.Server.Run: %v", err)
		}
	}
	return nil
}
