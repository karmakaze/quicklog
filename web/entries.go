package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/karmakaze/quicklog/storage"
)

type EntriesHandler struct {
	db *sql.DB
}

func NewEntriesHandler(db *sql.DB) *EntriesHandler {
	return &EntriesHandler{db: db}
}

func (h *EntriesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "OPTIONS":
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST")
	case "GET":
		h.listEntries(w, r)
	case "POST":
		h.createEntry(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *EntriesHandler) listEntries(w http.ResponseWriter, r *http.Request) {
	entries := make([]storage.Entry, 0)

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if tx, err := h.db.BeginTx(r.Context(), nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
	} else {
		err = storage.ListEntries("", "", &entries, tx, r.Context())
		if err != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
		}
		if data, err := json.Marshal(entries); err != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
		} else {
			tx.Commit()
			w.Write(data)
		}
	}
}

func (h *EntriesHandler) createEntry(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading POST /entries body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var entry storage.Entry
	if err = json.Unmarshal(body, &entry); err != nil {
		fmt.Printf("Error parsing POST /entries body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// create the entry
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
		return
	}

	if err = storage.CreateEntry(entry, tx, r.Context()); err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
		return
	}

	tx.Commit()
	w.WriteHeader(http.StatusCreated)
}
