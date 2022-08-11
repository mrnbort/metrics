package storage

import (
	"github.com/umputun/metrics/metric"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

type DBAccessor struct {
	db []metric.Entry
}

func (d *DBAccessor) Write(m metric.Entry) error {
	d.db = append(d.db, m)
	return nil
}

func (d *DBAccessor) Delete(m metric.Entry) error {
	for i, v := range d.db {
		if v.Name != m.Name {
			continue
		}
		d.db = append(d.db[:i], d.db[i+1:]...)
	}
	return nil
}
