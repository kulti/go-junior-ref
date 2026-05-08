//go:build integration

//nolint:paralleltest //database tests use single database instance
package store_test

import (
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"

	"github.com/kulti/task_list_course/internal/app/apiserver/internal/models"
	"github.com/kulti/task_list_course/internal/app/apiserver/internal/store"
	"github.com/kulti/task_list_course/internal/pgconn"
)

func TestCreateList(t *testing.T) {
	store := setupStore(t)

	list := models.List{
		ID:    faker.UUIDHyphenated(),
		Name:  faker.Name(),
		Items: []models.Item{},
	}
	require.NoError(t, store.CreateList(t.Context(), list))

	checkList(t, store, list)
}

func TestCreateListWithItems(t *testing.T) {
	store := setupStore(t)

	list := models.List{
		ID:   faker.UUIDHyphenated(),
		Name: faker.Name(),
		Items: []models.Item{
			{
				ID:   faker.UUIDHyphenated(),
				Name: faker.Name(),
			},
		},
	}
	require.NoError(t, store.CreateList(t.Context(), list))

	checkList(t, store, models.List{
		ID:    list.ID,
		Name:  list.Name,
		Items: []models.Item{},
	})

	list.Items[0].ListID = list.ID
	require.NoError(t, store.CreateItem(t.Context(), list.Items[0]))
	checkList(t, store, list)
}

func checkList(t *testing.T, store *store.Store, list models.List) {
	t.Helper()

	actList, err := store.GetList(t.Context(), list.ID)
	require.NoError(t, err)
	require.Equal(t, list, actList)
}

func setupStore(t *testing.T) *store.Store {
	t.Helper()

	conn, err := pgconn.New(pgconn.Params{
		ConnectionString: "postgres://postgres:password@127.0.0.1:5432/api_server_db?sslmode=disable",
		ChecknInterval:   time.Second,
	})
	require.NoError(t, err)

	go func() {
		conn.Run(t.Context())
	}()

	require.Eventually(t, func() bool {
		return conn.Ready() == nil
	}, 10*time.Second, 100*time.Millisecond, "pgconn not ready in time")

	return store.New(store.Params{PgConn: conn})
}
