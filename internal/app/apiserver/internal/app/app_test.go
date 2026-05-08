package app_test

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"

	"github.com/kulti/task_list_course/internal/app/apiserver/internal/models"
)

//go:generate mockgen -destination mock_test.go -source=app.go -package=app_test -mock_names store=MockStore,publisher=MockPublisher

func TestSubscribe(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	listID := faker.UUIDHyphenated()
	email := faker.Email()
	f.ExpectStoreSubscribe(listID, email)

	f.Subscribe(context.Background(), t, listID, email)
}

func TestGetList(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	list := models.List{
		ID:   faker.UUIDHyphenated(),
		Name: faker.Word(),
	}
	f.ExpectStoreGetList(list)

	retList := f.GetList(context.Background(), t, list.ID)
	require.Equal(t, list, retList)
}

func TestCreateList(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	listName := faker.Word()
	listID := f.ExpectStoreCreateList(listName)

	retListID := f.CreateList(context.Background(), t, listName)
	require.Equal(t, *listID, retListID)
}
