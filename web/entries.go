package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
		respondNoContent(w)
	case "GET":
		h.listEntries(w, r)
	case "POST":
		h.createEntry(w, r)
	case "DELETE":
		h.deleteEntries(w, r)
	default:
		respondStatus(http.StatusMethodNotAllowed, w)
	}
}

func (h *EntriesHandler) listEntries(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

    projectId := 0
    if r.Referer() != "" {
        if refererUrl, err := url.Parse(r.Referer()); err == nil {
            projects := make([]storage.Project, 0, 1)
            if err := storage.ListProjects("domain", refererUrl.Hostname(), &projects, h.db, r.Context()); err == nil {
                for _, p := range projects {
                    projectId = int(p.Id)
                }
            }
        } else {
            log.Printf("Could not parse 'referer' URL: %s\n'", r.Referer())
        }
    }

    if projectId == 0 {
        var err error
        if projectId, err = strconv.Atoi(r.FormValue("project_id")); err != nil {
            badRequest("'project_id' is required (numeric)", w)
            return
        }
    }

	seqMin, seqMax, ok := parseIntRange("seq", r)

	if !ok {
		badRequest("'seq' must be 'from,' or ',to' or 'from,to' (integer values)", w)
		return
	}
	publishedMin, publishedMax, ok := parseTimeRange("published", r)
	if !ok {
		badRequest("'published' must be 'from,' or ',to' or 'from,to' in RFC 3339 format", w)
		return
	}
	if (seqMin != storage.MinInt || seqMax != storage.MaxInt) && (!publishedMin.IsZero() || !publishedMax.IsZero()) {
		badRequest("'seq' and 'published' cannot both be specified", w)
		return
	}

	traceId := r.FormValue("trace_id")
	spanId := r.FormValue("span_id")
	search := r.FormValue("search")
	tag := r.FormValue("tag")
	if search != "" && tag != "" {
		badRequest("'tag' cannot be specified with 'search'", w)
		return
	}
	if tag != "" {
		search = "tag:" + search
	}
	if (traceId != "" || spanId != "") && (search != "" || tag != "") {
		badRequest("'trace_id' or 'span_id' cannot be specified with 'search'/'tag'", w)
		return
	}
	if traceId != "" && spanId != "" && traceId != spanId {
		badRequest("'trace_id' and 'span_id' cannot both be specified (unless they are the same)", w)
		return
	}

	count := 100
	if value := r.FormValue("count"); value != "" {
	    var err error
		if count, err = strconv.Atoi(value); err != nil || count < 1 || count > 1000 {
			badRequest("'count' must be between 1 to 1000", w)
			return
		}
	}

	entries, err := storage.ListEntries(projectId, seqMin, seqMax, publishedMin, publishedMax,
		traceId, spanId, search, count, h.db, r.Context())
	if err != nil {
		badRequest(err.Error(), w)
		return
	}
	respondOK(entries, w)
}

func (h *EntriesHandler) createEntry(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		respondStatus(http.StatusUnsupportedMediaType, w)
		return
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		badRequest(fmt.Sprintf("Error reading POST /entries body: %v", err), w)
		return
	}

	var entry storage.Entry
	if err = json.Unmarshal(body, &entry); err != nil {
		badRequest(fmt.Sprintf("Error parsing POST /entries body: %v\n", err), w)
		return
	}

	if entry.ProjectId <= 0 {
		badRequest("'project_id' is required", w)
		return
	}

	entry.Seq = 0
	if entry.Published.IsZero() {
		badRequest("'published' is required", w)
		return
	}
	if entry.Source == "" {
		badRequest("'source' is required", w)
		return
	}
	if entry.Type == "" {
		badRequest("'type' is required", w)
		return
	}

	// create the entry

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		respondError(http.StatusInternalServerError, err, w)
		return
	}
	if err = storage.CreateEntry(entry, tx, r.Context()); err != nil {
		tx.Rollback()
		if !storage.IsUniqueViolation(err) {
			respondError(http.StatusInternalServerError, err, w)
			return
		}
	} else {
		tx.Commit()
	}
	respondCreated("", w)
}

func (h *EntriesHandler) deleteEntries(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	projectId, err := strconv.Atoi(r.FormValue("project_id"))
	if err != nil {
		badRequest("'project_id' is required", w)
		return
	}

	publishedMin, publishedMax, ok := parseTimeRange("published", r)
	if !ok || publishedMin.IsZero() && publishedMax.IsZero() {
		badRequest("'published' must be 'from,' or ',to' or 'from,to' in RFC 3339 format", w)
		return
	}

	log.Printf("Deleting entries: project_id: %d, published: %v\n", projectId, r.FormValue("published"))

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		respondError(http.StatusInternalServerError, err, w)
		return
	}

	if err = storage.DeleteEntries(projectId, publishedMin, publishedMax, tx, r.Context()); err != nil {
		tx.Rollback()
		respondError(http.StatusInternalServerError, err, w)
		return
	}
	tx.Commit()

	respondStatus(http.StatusNoContent, w)
}
