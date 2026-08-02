package handlers

import (
	"net/http"
	"strconv"
)

func parseLimitOffset(r *http.Request) (limit, offset int) {
	limit = parseIntQuery(r, "limit", 0)
	offset = parseIntQuery(r, "offset", 0)
	return limit, offset
}

func parseIntQuery(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func parseBoolQuery(r *http.Request, key string, def bool) bool {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func parseInt32Path(vars map[string]string, key string) (int32, bool) {
	v := vars[key]
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(n), true
}

func parseInt64Path(vars map[string]string, key string) (int64, bool) {
	v := vars[key]
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func parseOptionalInt32Query(r *http.Request, key string) *int32 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return nil
	}
	i := int32(n)
	return &i
}

func parseOptionalFloatQuery(r *http.Request, key string) *float64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return &f
}

func parseOptionalBoolQuery(r *http.Request, key string) *bool {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}

func writeJSONArray(w http.ResponseWriter, status int, payload any) {
	if payload == nil {
		payload = []any{}
	}
	writeJSON(w, status, payload)
}
