package web

import (
	"net/http"
	"strings"
	"time"
)

func parsePublished(r *http.Request) (time.Time, time.Time, bool) {
	var t0 time.Time
	publishedMin := t0
	publishedMax := t0

	if published := r.FormValue("published"); published != "" {
		if published[0] != '[' || published[len(published)-1] != ']' {
			return t0, t0, false
		}
		parts := strings.Split(published[1:len(published)-1], ",")
		if len(parts) != 2 {
			return t0, t0, false
		}

		rfc3339micro := "2006-01-02T15:04:05.999999Z07:00"
		var err error
		if parts[0] != "" {
			publishedMin, err = time.Parse(rfc3339micro, strings.Replace(parts[0], " ", "T", -1))
			if err != nil {
				return t0, t0, false
			}
		}
		if parts[1] != "" {
			publishedMax, err = time.Parse(rfc3339micro, strings.Replace(parts[1], " ", "T", -1))
			if err != nil {
				return t0, t0, false
			}
		}
	}
	return publishedMin, publishedMax, true
}
