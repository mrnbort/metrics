package metric

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

func TestDuration_UnmarshalJSON(t *testing.T) {
	type Message struct {
		Elapsed Duration `json:"elapsed"`
	}
	var msgs []time.Duration

	tbl := []struct {
		unmar []byte
		res   time.Duration
	}{
		{[]byte(`{"elapsed":"1h"}`),
			time.Hour,
		},
		{
			[]byte(`{"elapsed":"1m"}`),
			time.Minute,
		},
		{
			[]byte(`{"elapsed": 1800000000000}`),
			30 * time.Minute,
		},
	}

	for i, tt := range tbl {
		var msg Message
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			err := json.Unmarshal(tt.unmar, &msg)
			require.NoError(t, err)
			msgs = append(msgs, time.Duration(msg.Elapsed))
			assert.Equal(t, tt.res, msgs[i])
		})
	}
}
