package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umputun/metrics/metric"
	"strconv"
	"testing"
	"time"
)

func TestService_getMinSinceMidnight(t *testing.T) {
	tbl := []struct {
		tm  time.Time
		res int
	}{
		{
			time.Date(2022, time.July, 27, 0, 1, 23, 0, time.UTC),
			1,
		},
		{
			time.Date(2022, time.July, 27, 16, 23, 23, 0, time.UTC),
			983,
		},
	}

	svc := &Service{}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			assert.Equal(t, tt.res, svc.getMinSinceMidnight(tt.tm))
		})
	}
}

func TestService_doCleanup(t *testing.T) {
	db := &AccessorMock{
		WriteFunc: func(m metric.Entry) error {
			return nil
		},
	}

	svc := New(db)

	err := svc.Update(metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     3,
	})
	require.NoError(t, err)

	err = svc.Update(metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     4,
	})
	require.NoError(t, err)

	err = svc.Update(metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     11,
	})
	require.NoError(t, err)

	//svc.data check
	assert.Equal(t, 18, svc.data["file_1"].Value)

	err = svc.doCleanup()
	require.NoError(t, err)

	//svc.data check, some gone
	assert.Equal(t, 0, len(svc.data))

}

func TestService_Update(t *testing.T) {
	db := &AccessorMock{
		WriteFunc: func(m metric.Entry) error {
			return nil
		},
	}

	svc := New(db)

	err := svc.Update(metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     1,
	})
	require.NoError(t, err)

	err = svc.Update(metric.Entry{
		Name:      "file_2",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     2,
	})
	require.NoError(t, err)

	err = svc.Update(metric.Entry{
		Name:      "file_2",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     3,
	})
	require.NoError(t, err)

	assert.Equal(t, 1, svc.data["file_1"].Value)
	assert.Equal(t, 5, svc.data["file_2"].Value)

	err = svc.Update(metric.Entry{
		Name:      "file_2",
		TimeStamp: time.Date(2022, 7, 29, 12, 11, 23, 0, time.UTC),
		Value:     4,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, svc.data["file_1"].Value)
	assert.Equal(t, 4, svc.data["file_2"].Value)
}

func TestNew(t *testing.T) {
	db := &AccessorMock{
		WriteFunc: func(m metric.Entry) error {
			return nil
		},
	}

	svc := New(db)
	assert.Equal(t, 0, len(svc.data))
}

// Tests, interface, api

func TestService_Delete(t *testing.T) {
	db := &AccessorMock{
		DeleteFunc: func(m metric.Entry) error {
			return nil
		},
		WriteFunc: func(m metric.Entry) error {
			return nil
		},
	}

	svc := New(db)

	err := svc.Update(metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     1,
	})
	require.NoError(t, err)

	err = svc.Update(metric.Entry{
		Name:      "file_2",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     2,
	})
	require.NoError(t, err)

	err = svc.Update(metric.Entry{
		Name:      "file_2",
		TimeStamp: time.Date(2022, 7, 29, 12, 11, 23, 0, time.UTC),
		Value:     3,
	})
	require.NoError(t, err)

	err = svc.Delete(metric.Entry{
		Name: "file_2",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(svc.data))
	assert.Equal(t, 1, svc.data["file_1"].Value)
	// assert.Equal(t, 0, len(svc.db))
}