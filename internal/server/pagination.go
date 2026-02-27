package server

import (
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

const defaultLimit = 20
const maxLimit = 100

func parsePaginationParams(r *http.Request) (cursorSortOrder *int32, cursorID *string, limit int32) {
	limitStr := r.URL.Query().Get("limit")
	limit = defaultLimit
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= maxLimit && n <= math.MaxInt32 {
			limit = int32(n) //nolint:gosec // bounds checked above
		}
	}

	cursor := r.URL.Query().Get("cursor")
	if cursor == "" {
		return nil, nil, limit
	}

	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, nil, limit
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return nil, nil, limit
	}

	so, err := strconv.Atoi(parts[0])
	if err != nil || so < math.MinInt32 || so > math.MaxInt32 {
		return nil, nil, limit
	}

	sortOrder := int32(so) //nolint:gosec // bounds checked above
	return &sortOrder, &parts[1], limit
}

func encodeCursor(sortOrder int32, id string) string {
	raw := fmt.Sprintf("%d:%s", sortOrder, id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}
