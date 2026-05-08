package app_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kulti/task_list_course/internal/app/apiserver/internal/app"
	"github.com/kulti/task_list_course/internal/app/apiserver/internal/models"
)

type fixture struct {
	app    *app.App
	store  *MockStore
	pubish *MockPublisher
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	mockCtl := gomock.NewController(t)
	store := NewMockStore(mockCtl)
	publisher := NewMockPublisher(mockCtl)
	app := app.New(app.Params{
		Store:     store,
		Publisher: publisher,
	})
	return &fixture{
		app:    app,
		store:  store,
		pubish: publisher,
	}
}

func (f *fixture) ExpectStoreSubscribe(listID, email string) {
	f.store.EXPECT().Subscribe(gomock.Any(), listID, email).Return(nil)
}

func (f *fixture) ExpectStoreGetList(list models.List) {
	f.store.EXPECT().GetList(gomock.Any(), list.ID).Return(list, nil)
}

func (f *fixture) ExpectStoreCreateList(listName string) *string {
	var listID string
	f.store.EXPECT().CreateList(gomock.Any(), createListMatcher{listName}).
		Do(func(ctx context.Context, l models.List) {
			listID = l.ID
		})
	return &listID
}

func (f *fixture) Subscribe(ctx context.Context, t *testing.T, listID, email string) {
	t.Helper()
	err := f.app.Subscribe(ctx, listID, email)
	require.NoError(t, err)
}

func (f *fixture) GetList(ctx context.Context, t *testing.T, listID string) models.List {
	t.Helper()
	list, err := f.app.GetList(ctx, listID)
	require.NoError(t, err)
	return list
}

func (f *fixture) CreateList(ctx context.Context, t *testing.T, listName string) string {
	t.Helper()
	listID, err := f.app.CreateList(ctx, listName)
	require.NoError(t, err)
	return listID
}

type createListMatcher struct {
	name string
}

func (m createListMatcher) Matches(x any) bool {
	l, ok := x.(models.List)
	if !ok {
		return false
	}
	return l.Name == m.name
}

func (m createListMatcher) String() string {
	return fmt.Sprintf("models.List with Name=%s", m.name)
}

func (m createListMatcher) Got(x any) string {
	return fmt.Sprintf("%#v", x)
}
