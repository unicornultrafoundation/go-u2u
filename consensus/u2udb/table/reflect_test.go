package table

//go:generate go run github.com/golang/mock/mockgen -package=table -destination=mock_test.go github.com/Fantom-foundation/lachesis-base/u2udb DBProducer,DropableStore

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	u2udb "github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
)

type testTables struct {
	NoTable interface{}
	Manual  u2udb.Store `table:"-"`
	Nil     u2udb.Store `table:"-"`
	Auto1   u2udb.Store `table:"A"`
	Auto2   u2udb.Store `table:"B"`
	Auto3   u2udb.Store `table:"C"`
}

func TestOpenTables(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	prefix := "prefix"

	mockStore := func() u2udb.Store {
		store := NewMockDropableStore(ctrl)
		store.EXPECT().Close().
			Times(1).
			Return(nil)
		return store
	}

	dbs := NewMockDBProducer(ctrl)
	dbs.EXPECT().OpenDB(gomock.Any()).
		Times(3).
		DoAndReturn(func(name string) (u2udb.Store, error) {
			require.Contains(name, prefix)
			return mockStore(), nil
		})

	tt := &testTables{}

	// open auto
	err := OpenTables(tt, dbs, prefix)
	require.NoError(err)
	require.NotNil(tt.Auto1)
	require.NotNil(tt.Auto2)
	require.NotNil(tt.Auto3)
	require.Nil(tt.NoTable)
	require.Nil(tt.Nil)

	// open manual
	require.Nil(tt.Manual)
	tt.Manual = mockStore()
	require.NotNil(tt.Manual)

	// close all
	err = CloseTables(tt)
	require.NoError(err)
	require.NotNil(tt.Auto1)
	require.NotNil(tt.Auto2)
	require.NotNil(tt.Auto3)
	require.Nil(tt.NoTable)
	require.Nil(tt.Nil)
	require.NotNil(tt.Manual)
}
