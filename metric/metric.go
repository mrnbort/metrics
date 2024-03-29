package metric

import (
	"encoding/json"
	"fmt"
	"time"
)

// Entry creates a metric to save/delete from the db
type Entry struct {
	Name      string    `bson:"name" json:"name"`
	TimeStamp time.Time `bson:"time_stamp" json:"time_stamp"`
	Value     int       `bson:"value" json:"value"`

	MinSinceMidnight int           `bson:"-" json:"-"`
	Type             time.Duration `bson:"type" json:"type"`
	TypeStr          string        `bson:"type_str" json:"type_str"`
}

// Lookup criteria for metric/metrics in db
type Lookup struct {
	Name     string    `json:"name"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Interval Duration  `json:"interval"`
}

// Duration custom type
type Duration time.Duration

// UnmarshalJSON to unmarshal json that is either float or string to time.Duration
func (duration *Duration) UnmarshalJSON(b []byte) error {
	var unmarshalledJson interface{}

	err := json.Unmarshal(b, &unmarshalledJson)
	if err != nil {
		return err
	}

	switch value := unmarshalledJson.(type) {
	case float64:
		*duration = Duration(time.Duration(value))
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*duration = Duration(tmp)
	default:
		return fmt.Errorf("invalid duration: %#v", unmarshalledJson)
	}
	return nil
}
