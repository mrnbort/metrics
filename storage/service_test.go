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

	err := svc.Update(metric.Entry{})
	require.NoError(t, err)

	err = svc.Update(metric.Entry{})
	require.NoError(t, err)

	err = svc.Update(metric.Entry{})
	require.NoError(t, err)

	//svc.data check

	err = svc.doCleanup()
	require.NoError(t, err)

	//svc.data check, some gone

}
