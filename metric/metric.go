package metric

import (
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
	Interval string    `json:"interval"`
}
