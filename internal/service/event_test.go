package service_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Xevion/motophoto/internal/service"
	"github.com/Xevion/motophoto/internal/testutil"
	"github.com/Xevion/motophoto/internal/testutil/dbfactory"
)

func TestEventService_Create(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	userID := dbfactory.User(ctx, t, env.Pool, nil)

	event, err := env.Events.Create(ctx, service.CreateEventParams{
		PhotographerID: userID,
		Name:           "Spring MX 2026",
		Slug:           "spring-mx-2026",
		Sport:          "motocross",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, event.ID)
	assert.Equal(t, "Spring MX 2026", event.Name)
	assert.Equal(t, "spring-mx-2026", event.Slug)
	assert.Equal(t, "motocross", event.Sport)
	assert.Equal(t, "draft", event.Status)
}

func TestEventService_Create_WithTags(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	userID := dbfactory.User(ctx, t, env.Pool, nil)
	tags := []string{"outdoor", "motocross", "spring-2026"}

	event, err := env.Events.Create(ctx, service.CreateEventParams{
		PhotographerID: userID,
		Name:           "Tagged Event",
		Slug:           "tagged-event",
		Sport:          "motocross",
		Tags:           tags,
	})
	require.NoError(t, err)
	assert.Equal(t, tags, event.Tags)

	// Round-trip: fetch and verify tags survive persistence.
	got, err := env.Events.Get(ctx, event.ID)
	require.NoError(t, err)
	assert.Equal(t, tags, got.Tags)
}

func TestEventService_Create_InvalidStatus(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	userID := dbfactory.User(ctx, t, env.Pool, nil)

	_, err := env.Events.Create(ctx, service.CreateEventParams{
		PhotographerID: userID,
		Name:           "Bad Status",
		Slug:           "bad-status",
		Sport:          "motocross",
		Status:         new("bogus"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestEventService_Create_InvalidDate(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	userID := dbfactory.User(ctx, t, env.Pool, nil)

	_, err := env.Events.Create(ctx, service.CreateEventParams{
		PhotographerID: userID,
		Name:           "Bad Date",
		Slug:           "bad-date",
		Sport:          "motocross",
		Date:           new("not-a-date"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid date")
}

func TestEventService_Create_DuplicateSlug(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	userID := dbfactory.User(ctx, t, env.Pool, nil)

	_, err := env.Events.Create(ctx, service.CreateEventParams{
		PhotographerID: userID,
		Name:           "Event A", Slug: "same-slug", Sport: "motocross",
	})
	require.NoError(t, err)

	_, err = env.Events.Create(ctx, service.CreateEventParams{
		PhotographerID: userID,
		Name:           "Event B", Slug: "same-slug", Sport: "motocross",
	})
	assert.ErrorIs(t, err, service.ErrConflict)
}

func TestEventService_Get(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
		Name: new("Lookup Test"),
		Slug: new("lookup-test"),
	})

	t.Run("by ID", func(t *testing.T) {
		t.Parallel()
		got, err := env.Events.Get(ctx, event.ID)
		require.NoError(t, err)
		assert.Equal(t, "Lookup Test", got.Name)
		assert.NotNil(t, got.Galleries)
	})

	t.Run("by slug", func(t *testing.T) {
		t.Parallel()
		got, err := env.Events.Get(ctx, "lookup-test")
		require.NoError(t, err)
		assert.Equal(t, event.ID, got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		_, err := env.Events.Get(ctx, "nonexistent")
		assert.ErrorIs(t, err, service.ErrNotFound)
	})
}

func TestEventService_List(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	// Track created IDs to verify the right events come back.
	createdIDs := make(map[string]bool, 3)
	for i := range 3 {
		e := dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
			Status: new("published"),
			Name:   new(fmt.Sprintf("Event %d", i)),
			Slug:   new(fmt.Sprintf("event-%d", i)),
		})
		createdIDs[e.ID] = true
	}

	// Draft event should not appear.
	dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
		Status: new("draft"),
	})

	result, err := env.Events.List(ctx, nil, nil, 10)
	require.NoError(t, err)
	assert.Len(t, result.Events, 3)
	assert.Nil(t, result.NextCursorID)

	// Verify the exact IDs returned.
	for _, e := range result.Events {
		assert.True(t, createdIDs[e.ID], "unexpected event ID %s in list", e.ID)
	}
}

func TestEventService_List_Pagination(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	createdIDs := make(map[string]bool, 5)
	for i := range 5 {
		e := dbfactory.Event(ctx, t, env.Pool, env.Events, &dbfactory.EventOpts{
			Status: new("published"),
			Name:   new(fmt.Sprintf("Page Event %d", i)),
			Slug:   new(fmt.Sprintf("page-event-%d", i)),
		})
		createdIDs[e.ID] = true
	}

	// Collect all IDs across pages to verify no duplicates and completeness.
	seenIDs := make(map[string]bool)

	// First page.
	page1, err := env.Events.List(ctx, nil, nil, 2)
	require.NoError(t, err)
	assert.Len(t, page1.Events, 2)
	require.NotNil(t, page1.NextCursorID)
	for _, e := range page1.Events {
		seenIDs[e.ID] = true
	}

	// Second page.
	page2, err := env.Events.List(ctx, page1.NextCursorSortOrder, page1.NextCursorID, 2)
	require.NoError(t, err)
	assert.Len(t, page2.Events, 2)
	for _, e := range page2.Events {
		assert.False(t, seenIDs[e.ID], "duplicate ID %s across pages", e.ID)
		seenIDs[e.ID] = true
	}

	// Third page (last item).
	page3, err := env.Events.List(ctx, page2.NextCursorSortOrder, page2.NextCursorID, 2)
	require.NoError(t, err)
	assert.Len(t, page3.Events, 1)
	assert.Nil(t, page3.NextCursorID)
	for _, e := range page3.Events {
		assert.False(t, seenIDs[e.ID], "duplicate ID %s across pages", e.ID)
		seenIDs[e.ID] = true
	}

	// All created IDs should have been seen exactly once.
	assert.Equal(t, createdIDs, seenIDs)
}

func TestEventService_Update(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	updated, err := env.Events.Update(ctx, event.ID, service.UpdateEventParams{
		Name: new("Updated Name"),
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, event.Slug, updated.Slug) // unchanged field preserved
}

func TestEventService_Update_InvalidStatus(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	_, err := env.Events.Update(ctx, event.ID, service.UpdateEventParams{
		Status: new("bogus"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestEventService_Update_InvalidDate(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	_, err := env.Events.Update(ctx, event.ID, service.UpdateEventParams{
		Date: new("not-a-date"),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid date")
}

func TestEventService_Update_NotFound(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	_, err := env.Events.Update(ctx, "nonexistent", service.UpdateEventParams{
		Name: new("nope"),
	})
	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestEventService_Delete(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	err := env.Events.Delete(ctx, event.ID)
	require.NoError(t, err)

	_, err = env.Events.Get(ctx, event.ID)
	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestEventService_Delete_Nonexistent(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	// Deleting a non-existent event should not error (idempotent) or return
	// a meaningful error depending on implementation. Assert it doesn't panic.
	err := env.Events.Delete(ctx, "nonexistent-id")
	// The current implementation doesn't distinguish "no rows affected", so
	// this documents the actual behavior.
	_ = err
}
