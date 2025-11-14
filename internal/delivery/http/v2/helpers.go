package v2

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"ppo/internal/delivery/http/dto"
)

func decodeJSON(r *http.Request, v interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return &dto.BadRequestError{Message: "invalid JSON payload"}
	}
	return nil
}

func parsePagination(r *http.Request, defaultLimit, maxLimit int) (limit int, offset int, err error) {
	values := r.URL.Query()
	limit, err = parsePositiveInt(values, "limit", defaultLimit, 1, maxLimit)
	if err != nil {
		return 0, 0, err
	}
	offset, err = parsePositiveInt(values, "offset", 0, 0, 10_000_000)
	if err != nil {
		return 0, 0, err
	}
	return limit, offset, nil
}

func parsePositiveInt(values url.Values, key string, def, min, max int) (int, error) {
	raw := values.Get(key)
	if raw == "" {
		return def, nil
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return 0, &dto.BadRequestError{Message: key + " must be a number"}
	}
	if val < min {
		return 0, &dto.BadRequestError{Message: key + " is below minimum"}
	}
	if val > max {
		return max, nil
	}
	return val, nil
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func collectMultipartTags(form *multipart.Form, keys ...string) string {
	if form == nil {
		return ""
	}
	seen := make(map[string]struct{})
	var tags []string
	for _, key := range keys {
		values := form.Value[key]
		for _, raw := range values {
			for _, part := range strings.Split(raw, ",") {
				trimmed := strings.TrimSpace(part)
				if trimmed == "" {
					continue
				}
				if _, ok := seen[trimmed]; ok {
					continue
				}
				seen[trimmed] = struct{}{}
				tags = append(tags, trimmed)
			}
		}
	}
	if len(tags) == 0 {
		return ""
	}
	return strings.Join(tags, ",")
}
