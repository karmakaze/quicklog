package web

import (
	"log"
	"net/http"
	"strconv"
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

func Serve(port int, dbUrl string) error {
    db, err := storage.OpenDB(dbUrl)
    if db != nil {
        defer db.Close()
    }
    if err != nil {
        return err
    }

	// these get added to http.DefaultServeMux
	http.Handle("/projects", NewProjectsHandler(db))
	http.Handle("/entries", NewEntriesHandler(db))
	http.Handle("/tags", NewTagsHandler(db))

    log.Printf("Listening on port %d\n", port)
    return http.ListenAndServe(":" + strconv.Itoa(port), nil)
}
