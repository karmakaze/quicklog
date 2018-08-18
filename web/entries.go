package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	case "DELETE":
		h.deleteEntries(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *EntriesHandler) listEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	r.ParseForm()

	projectId, err := strconv.Atoi(r.FormValue("project_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "'project_id' is required (numeric)"}`))
		return
	}

	publishedMin, publishedMax, ok := parsePublished(r)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "'published' must be [from,] or [,to] or [from,to] in RFC 3339 format"}`))
		return
	}

	count := 100
	if value := r.FormValue("count"); value != "" {
		if count, err = strconv.Atoi(value); err != nil || count < 1 || count > 1000 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "'count' must be between 1 to 1000"}`))
			return
		}
	}

	if tx, err := h.db.BeginTx(r.Context(), nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
	} else {
		entries, err := storage.ListEntries(projectId, publishedMin, publishedMax, count, tx, r.Context())
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

	if entry.ProjectId <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "'project_id' is required"}`))
	}
	entry.Seq = 0
	if entry.Published.IsZero() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "'published' is required"}`))
	}
	if entry.Source == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "'source' is required"}`))
	}
	if entry.Type == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "'type' is required"}`))
	}

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

func (h *EntriesHandler) deleteEntries(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	projectId, err := strconv.Atoi(r.FormValue("project_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "'project_id' is required"}`))
		return
	}

	publishedMin, publishedMax, ok := parsePublished(r)
	if !ok || publishedMin.IsZero() && publishedMax.IsZero() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "'published' must be [from,] or [,to] or [from,to] in RFC 3339 format"}`))
		return
	}

	if apiKey := r.FormValue("api_key"); apiKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "'api_key' is required"}`))
		return
	} else {
		if !storage.VerifyApiKey(projectId, apiKey, h.db, r.Context()) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "'api_key' is not valid"}`))
			return
		}
	}

	log.Printf("Deleting entries: project_id: %d, published: %v\n", projectId, r.FormValue("published"))

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
		return
	}

	if err = storage.DeleteEntries(projectId, publishedMin, publishedMax, tx, r.Context()); err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
		return
	}
	tx.Commit()

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
