package storage

import (
	"context"
	"fmt"
	"github.com/umputun/metrics/metric"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

//go:generate moq -out dbaccessor_mock.go . DBAccessor

// DBAccessor initiates MongoDB
type DBAccessor struct {
	//db []metric.Entry
	db               *mongo.Client
	dbName, collName string
}

func NewAccessor(db *mongo.Client, dbName, collName string) *DBAccessor {
	return &DBAccessor{db: db, dbName: dbName, collName: collName}
}

// Write inserts entries to db
func (d *DBAccessor) Write(ctx context.Context, m metric.Entry) error {
	collection := d.db.Database(d.dbName).Collection(d.collName)
	if _, err := collection.InsertOne(ctx, m); err != nil {
		return fmt.Errorf("failed it write %+v: %w", m, err)
	}
	log.Printf("inserted metric: %v", m.Name)
	return nil
}

// Delete removes entries from db
func (d *DBAccessor) Delete(m metric.Entry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := d.db.Database("metrics-service").Collection("metrics")

	_, err := collection.DeleteMany(ctx, m)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("deleted metric %v\n", m.Name)
	return nil
}

// FindAll gets all entries for the specified timeframe and interval from db
func (d *DBAccessor) FindAll(from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := d.db.Database("metrics-service").Collection("metrics")

	var metrics []metric.Entry

	cursor, err := collection.Find(ctx, bson.M{"time_stamp": bson.M{
		"$gt": from,
		"$lt": to,
	},
	})
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(ctx, &metrics); err != nil {
		log.Fatal(err)
	}

	return metrics, nil
}
