package metric

import "time"

type Entry struct {
	Name      string    `bson:"name" json:"name"`
	TimeStamp time.Time `bson:"time_stamp" json:"time_stamp"`
	Value     int       `bson:"value" json:"value"`

	MinSinceMidnight int `json:"-" bson:"-"`
}