package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/karmakaze/quicklog/storage"
)

type ProjectsHandler struct {
	db *sql.DB
}

func NewProjectsHandler(db *sql.DB) *ProjectsHandler {
	return &ProjectsHandler{db: db}
}

func (h *ProjectsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "OPTIONS":
		respondNoContent(w)
	case "GET":
		h.listProjects(w, r)
	case "POST":
		h.createProject(w, r)
	default:
		respondStatus(http.StatusMethodNotAllowed, w)
	}
}

func (h *ProjectsHandler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects := make([]storage.Project, 0)

	err := storage.ListProjects("", "", &projects, h.db, r.Context())
	if err != nil {
		badRequest(err.Error(), w)
		return
	}
	respondOK(projects, w)
}

func (h *ProjectsHandler) createProject(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		respondStatus(http.StatusUnsupportedMediaType, w)
		return
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		badRequest(fmt.Sprintf("Error reading POST /projects body: %v\n", err), w)
		return
	}

	var project storage.Project
	if err = json.Unmarshal(body, &project); err != nil {
		badRequest(fmt.Sprintf("Error parsing POST /projects body: %v\n", err), w)
		return
	}

	project.Id = 0
	if project.Name == "" {
		badRequest("'name' is required", w)
		return
	}

	// create the entry

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		respondError(http.StatusInternalServerError, err, w)
		return
	}

	if err = storage.CreateProject(project, tx, r.Context()); err != nil {
		tx.Rollback()
		respondError(http.StatusInternalServerError, err, w)
		return
	}

	tx.Commit()
	respondCreated("", w)
}
