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
	db               *mongo.Client
	dbName, collName string
}

//
func NewAccessor(db *mongo.Client, dbName, collName string) *DBAccessor {
	return &DBAccessor{db: db, dbName: dbName, collName: collName}
}

// Write inserts entries to db
func (d *DBAccessor) Write(ctx context.Context, m metric.Entry) error {
	collection := d.db.Database(d.dbName).Collection(d.collName)
	m.Type = "1m"
	if _, err := collection.InsertOne(ctx, m); err != nil {
		return fmt.Errorf("failed to write %+v: %w", m, err)
	}
	log.Printf("inserted metric: %v", m.Name)
	return nil
}

// Delete removes entries from db
func (d *DBAccessor) Delete(ctx context.Context, m metric.Entry) error {
	collection := d.db.Database(d.dbName).Collection(d.collName)
	if _, err := collection.DeleteMany(ctx, m); err != nil {
		return fmt.Errorf("failed to delete %v: %w", m, err)
	}
	fmt.Printf("deleted metric %v\n", m.Name)
	return nil
}

// GetMetricsList gets a list of available metrics in db
func (d *DBAccessor) GetMetricsList(ctx context.Context) ([]string, error) {
	var metricsList []string

	collection := d.db.Database(d.dbName).Collection(d.collName)
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return metricsList, fmt.Errorf("failed to read all documents: %w", err)
	}

	defer cursor.Close(ctx) // ???????

	for cursor.Next(ctx) {
		var result metric.Entry
		if err := cursor.Decode(&result); err != nil {
			return metricsList, fmt.Errorf("failed to decode from db: %w", err)
		}
		metricsList = append(metricsList, result.Name)

	}
	return metricsList, nil
}

// FindOneMetric gets the values for the metric, timeframe and interval from db
func (d *DBAccessor) FindOneMetric(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	var results []metric.Entry
	return results, nil
}

// FindAll gets all entries for the specified timeframe and interval from db
func (d *DBAccessor) FindAll(ctx context.Context, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	var metrics []metric.Entry

	collection := d.db.Database(d.dbName).Collection(d.collName)
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
