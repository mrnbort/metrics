package storage

import (
	"context"
	"fmt"
	"github.com/umputun/metrics/metric"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

//go:generate moq -out dbaccessor_mock.go . DBAccessor

type DBAccessor struct {
	//db []metric.Entry
	db *mongo.Client
}

func (d *DBAccessor) Write(m metric.Entry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	d.db, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	collection := d.db.Database("metrics-service").Collection("metrics")

	_, err := collection.InsertOne(ctx, m)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("inserted metric %v\n", m.Name)
	return nil
}

func (d *DBAccessor) Delete(m metric.Entry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	d.db, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	collection := d.db.Database("metrics-service").Collection("metrics")

	_, err := collection.DeleteMany(ctx, m)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("deleted metric %v\n", m.Name)
	return nil
}

func (d *DBAccessor) FindAll(from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	d.db, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
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
