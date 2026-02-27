package server

import (
	"encoding/base64"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePaginationParams_DefaultLimit(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	sortOrder, id, limit := parsePaginationParams(r)
	assert.Nil(t, sortOrder)
	assert.Nil(t, id)
	assert.Equal(t, int32(defaultLimit), limit)
}

func TestParsePaginationParams_ValidLimit(t *testing.T) {
	r := httptest.NewRequest("GET", "/?limit=50", nil)
	_, _, limit := parsePaginationParams(r)
	assert.Equal(t, int32(50), limit)
}

func TestParsePaginationParams_LimitExceedsMax(t *testing.T) {
	r := httptest.NewRequest("GET", "/?limit=9999", nil)
	_, _, limit := parsePaginationParams(r)
	assert.Equal(t, int32(defaultLimit), limit)
}

func TestParsePaginationParams_InvalidLimitString(t *testing.T) {
	r := httptest.NewRequest("GET", "/?limit=abc", nil)
	_, _, limit := parsePaginationParams(r)
	assert.Equal(t, int32(defaultLimit), limit)
}

func TestParsePaginationParams_ValidCursor(t *testing.T) {
	cursor := base64.RawURLEncoding.EncodeToString([]byte("42:some-uuid"))
	r := httptest.NewRequest("GET", "/?cursor="+cursor, nil)
	sortOrder, id, limit := parsePaginationParams(r)
	assert.NotNil(t, sortOrder)
	assert.Equal(t, int32(42), *sortOrder)
	assert.NotNil(t, id)
	assert.Equal(t, "some-uuid", *id)
	assert.Equal(t, int32(defaultLimit), limit)
}

func TestParsePaginationParams_InvalidCursorEncoding(t *testing.T) {
	r := httptest.NewRequest("GET", "/?cursor=not-valid-base64!!!", nil)
	sortOrder, id, limit := parsePaginationParams(r)
	assert.Nil(t, sortOrder)
	assert.Nil(t, id)
	assert.Equal(t, int32(defaultLimit), limit)
}

func TestParsePaginationParams_CursorMissingColon(t *testing.T) {
	cursor := base64.RawURLEncoding.EncodeToString([]byte("nocolon"))
	r := httptest.NewRequest("GET", "/?cursor="+cursor, nil)
	sortOrder, id, _ := parsePaginationParams(r)
	assert.Nil(t, sortOrder)
	assert.Nil(t, id)
}

func TestParsePaginationParams_CursorInvalidSortOrder(t *testing.T) {
	cursor := base64.RawURLEncoding.EncodeToString([]byte("notanint:some-uuid"))
	r := httptest.NewRequest("GET", "/?cursor="+cursor, nil)
	sortOrder, id, _ := parsePaginationParams(r)
	assert.Nil(t, sortOrder)
	assert.Nil(t, id)
}

func TestEncodeCursor(t *testing.T) {
	id := "abc-123"
	sortOrder := int32(7)
	encoded := encodeCursor(sortOrder, id)
	expected := base64.RawURLEncoding.EncodeToString(fmt.Appendf(nil, "%d:%s", sortOrder, id))
	assert.Equal(t, expected, encoded)
}

func TestEncodeCursorRoundtrip(t *testing.T) {
	id := "round-trip-id"
	sortOrder := int32(99)
	cursor := encodeCursor(sortOrder, id)

	r := httptest.NewRequest("GET", "/?cursor="+cursor, nil)
	gotSortOrder, gotID, _ := parsePaginationParams(r)

	assert.NotNil(t, gotSortOrder)
	assert.Equal(t, sortOrder, *gotSortOrder)
	assert.NotNil(t, gotID)
	assert.Equal(t, id, *gotID)
}
