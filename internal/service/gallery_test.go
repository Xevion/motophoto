package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Xevion/motophoto/internal/service"
	"github.com/Xevion/motophoto/internal/testutil"
	"github.com/Xevion/motophoto/internal/testutil/dbfactory"
)

func TestGalleryService_Create(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	gal, err := env.Galleries.Create(ctx, event.ID, service.CreateGalleryParams{
		Name: "Race Day Photos",
		Slug: "race-day-photos",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, gal.ID)
	assert.Equal(t, "Race Day Photos", gal.Name)
	assert.Equal(t, "race-day-photos", gal.Slug)
}

func TestGalleryService_Create_DuplicateSlug(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	_, err := env.Galleries.Create(ctx, event.ID, service.CreateGalleryParams{
		Name: "Gallery A", Slug: "same-slug",
	})
	require.NoError(t, err)

	_, err = env.Galleries.Create(ctx, event.ID, service.CreateGalleryParams{
		Name: "Gallery B", Slug: "same-slug",
	})
	assert.ErrorIs(t, err, service.ErrConflict)
}

func TestGalleryService_Create_EventNotFound(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	_, err := env.Galleries.Create(ctx, "nonexistent-event", service.CreateGalleryParams{
		Name: "Gallery", Slug: "gallery",
	})
	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestGalleryService_List(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	g1 := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)
	g2 := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)

	galleries, err := env.Galleries.List(ctx, event.ID)
	require.NoError(t, err)
	assert.Len(t, galleries, 2)

	// Verify exact IDs returned.
	ids := map[string]bool{galleries[0].ID: true, galleries[1].ID: true}
	assert.True(t, ids[g1.ID], "expected gallery %s in list", g1.ID)
	assert.True(t, ids[g2.ID], "expected gallery %s in list", g2.ID)
}

func TestGalleryService_List_EventNotFound(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	_, err := env.Galleries.List(ctx, "nonexistent-event")
	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestGalleryService_Update(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)
	gal := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)

	updated, err := env.Galleries.Update(ctx, event.ID, gal.ID, service.UpdateGalleryParams{
		Name: new("Updated Gallery"),
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Gallery", updated.Name)
	assert.Equal(t, gal.Slug, updated.Slug)
}

func TestGalleryService_Update_NotFound(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	_, err := env.Galleries.Update(ctx, event.ID, "nonexistent", service.UpdateGalleryParams{
		Name: new("nope"),
	})
	assert.ErrorIs(t, err, service.ErrNotFound)
}

func TestGalleryService_Delete(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)
	gal := dbfactory.Gallery(ctx, t, env.Galleries, event.ID, nil)

	err := env.Galleries.Delete(ctx, event.ID, gal.ID)
	require.NoError(t, err)

	galleries, err := env.Galleries.List(ctx, event.ID)
	require.NoError(t, err)
	assert.Empty(t, galleries)
}

func TestGalleryService_Delete_Nonexistent(t *testing.T) {
	t.Parallel()
	env := testutil.NewEnv(t)
	ctx := t.Context()

	event := dbfactory.Event(ctx, t, env.Pool, env.Events, nil)

	// Documents behavior: deleting a non-existent gallery doesn't panic.
	err := env.Galleries.Delete(ctx, event.ID, "nonexistent-gallery")
	_ = err
}
