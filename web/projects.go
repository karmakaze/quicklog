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

type ProjectsHandler struct {
	db *sql.DB
}

func NewProjectsHandler(db *sql.DB) *ProjectsHandler {
	return &ProjectsHandler{db: db}
}

func (h *ProjectsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "OPTIONS":
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST")
	case "GET":
		h.listProjects(w, r)
	case "POST":
		h.createProject(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *ProjectsHandler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects := make([]storage.Project, 0)

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if tx, err := h.db.BeginTx(r.Context(), nil); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
	} else {
		err = storage.ListProjects("", "", &projects, tx, r.Context())
		if err != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
		}
		if data, err := json.Marshal(projects); err != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
		} else {
			tx.Commit()
			w.Write(data)
		}
	}
}

func (h *ProjectsHandler) createProject(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading POST /projects body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var project storage.Project
	if err = json.Unmarshal(body, &project); err != nil {
		fmt.Printf("Error parsing POST /projects body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// create the entry
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	project.Id = 0
	if project.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "'name' is required"}`))
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
		return
	}

	if err = storage.CreateProject(project, tx, r.Context()); err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": ` + strconv.Quote(err.Error()) + `}`))
		return
	}

	tx.Commit()
	w.WriteHeader(http.StatusCreated)
}
