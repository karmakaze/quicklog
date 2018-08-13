package web

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/karmakaze/quicklog/config"
	"github.com/karmakaze/quicklog/storage"
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
	db, err := storage.OpenDB(cfg)
	if err != nil {
		return err
	}

	http.HandleFunc("/status", status)
	http.Handle("/entries", NewEntriesHandler(db))

	certFile, privkeyFile := findCertAndPrivKey()
	if certFile != "" && privkeyFile != "" {
		log.Printf("Listening on %v:443 (with SSL)\n", addr)
		err := http.ListenAndServeTLS(addr+":443", certFile, privkeyFile, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	} else {
		log.Printf("Listening on %v:8124 (no-SSL)\n", addr)
		if err := http.ListenAndServe(addr+":8124", nil); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
