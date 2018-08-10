package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/karmakaze/quicklog/storage"
)

func parseIntRange(name string, r *http.Request) (int, int, bool) {
	if min, max, ok := parseRange(name, r); !ok {
		return 0, 0, false
	} else {
		iMin := storage.MinInt
		iMax := storage.MaxInt

		if min != "" {
			if i64, err := strconv.ParseInt(min, 10, 0); err != nil {
				return 0, 0, false
			} else {
				iMin = int(i64)
			}
		}
		if max != "" {
			if i64, err := strconv.ParseInt(max, 10, 0); err != nil {
				return 0, 0, false
			} else {
				iMax = int(i64)
			}
		}
		return iMin, iMax, true
	}
}

func parseTimeRange(name string, r *http.Request) (time.Time, time.Time, bool) {
	var t0 time.Time

	if min, max, ok := parseRange(name, r); !ok {
		return t0, t0, false
	} else {
		tMin := t0
		tMax := t0

		rfc3339micro := "2006-01-02T15:04:05.999999Z07:00"
		var err error
		if min != "" {
			tMin, err = time.Parse(rfc3339micro, strings.Replace(min, " ", "T", -1))
			if err != nil {
				return t0, t0, false
			}
		}
		if max != "" {
			tMax, err = time.Parse(rfc3339micro, strings.Replace(max, " ", "T", -1))
			if err != nil {
				return t0, t0, false
			}
		}
		return tMin, tMax, true
	}
}

func parseRange(name string, r *http.Request) (string, string, bool) {
	value := r.FormValue(name)
	value = strings.TrimLeft(value, "[(")
	value = strings.TrimRight(value, ")]")
	value = strings.Replace(value, "~", ",", -1)
	values := strings.Split(value, ",")

	if len(values) > 2 {
		return "", "", false
	}
	if strings.HasPrefix(value, ",") {
		return "", values[1], true
	} else if strings.HasSuffix(value, ",") {
		return values[0], "", true
	} else if len(values) == 2 {
		return values[0], values[1], true
	} else {
		return value, value, true
	}
}
