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
	"github.com/karmakaze/quicklog/storage/span_tag"
)

type Tag struct {
	ProjectId int32  `json:"project_id"`
	TraceId   string `json:"trace_id"`
	SpanId    string `json:"span_id"`
	Tag       string `json:"tag"`
}

type TagsHandler struct {
	db *sql.DB
}

func NewTagsHandler(db *sql.DB) *TagsHandler {
	return &TagsHandler{db: db}
}

func (h *TagsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "OPTIONS":
		respondNoContent(w)
	case "GET":
		h.listTags(w, r)
	case "POST":
		h.createTag(w, r)
	default:
		respondStatus(http.StatusMethodNotAllowed, w)
	}
}

func (h *TagsHandler) listTags(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	projectId, err := strconv.Atoi(r.FormValue("project_id"))
	if err != nil {
		badRequest("'project_id' is required (numeric)", w)
		return
	}

	traceId := r.FormValue("trace_id")
	spanId := r.FormValue("span_id")
	tag := r.FormValue("tag")

	if traceId == "" && spanId == "" && tag == "" {
		badRequest("'trace_id', 'span_id', or 'tag' must be specified", w)
		return
	}
	if (traceId != "" || spanId != "") && tag != "" {
		badRequest("'trace_id' or 'span_id' cannot be specified with 'tag'", w)
		return
	}
	if traceId != "" && spanId != "" && traceId != spanId {
		badRequest("'trace_id' and 'span_id' cannot both be specified (unless they are the same)", w)
		return
	}

	spanTags, err := span_tag.ListSpanTags(projectId, traceId, spanId, tag, h.db, r.Context())
	if err != nil {
		badRequest(err.Error(), w)
		return
	}
	respondOK(toTags(spanTags), w)
}

func (h *TagsHandler) createTag(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		respondStatus(http.StatusUnsupportedMediaType, w)
		return
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		badRequest(fmt.Sprintf("Error reading POST /tags body: %v", err), w)
		return
	}

	var tag Tag

	if err = json.Unmarshal(body, &tag); err != nil {
		badRequest(fmt.Sprintf("Error parsing POST /tags body: %v\n", err), w)
		return
	}

	if tag.ProjectId <= 0 {
		badRequest("'project_id' is required", w)
		return
	}

	if tag.TraceId == "" {
		badRequest("'trace_id' is required", w)
		return
	}

	if tag.SpanId == "" {
		badRequest("'span_id' is required", w)
		return
	}

	if tag.Tag == "" {
		badRequest("'tag' is required", w)
		return
	}

	// create the span_tag
	spanTag := toSpanTag(tag)

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		respondError(http.StatusInternalServerError, err, w)
		return
	}

	if err = span_tag.CreateSpanTag(spanTag, tx, r.Context()); err != nil {
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

func toSpanTag(tag Tag) span_tag.SpanTag {
	key, value := span_tag.ParseTag(tag.Tag)
	return span_tag.SpanTag{
		ProjectId: tag.ProjectId,
		TraceId:   tag.TraceId,
		SpanId:    tag.SpanId,
		Key:       key,
		Value:     value,
	}
}

func toTags(spanTags []span_tag.SpanTag) []Tag {
	tags := make([]Tag, len(spanTags))
	for i, spanTag := range spanTags {
		tag := spanTag.Key
		if spanTag.Value != "" {
			tag = spanTag.Key + ":" + spanTag.Value
		}
		tags[i] = Tag{
			ProjectId: spanTag.ProjectId,
			TraceId:   spanTag.TraceId,
			SpanId:    spanTag.SpanId,
			Tag:       tag,
		}
	}
	return tags
}
