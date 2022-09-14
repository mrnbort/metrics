package storage

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ReagrBucket struct {
	Interval time.Duration // 30m, 8h, 24h, 7d
	Age      time.Duration // 24h, 7d, ...
	SrcType  string
	DstType  string
}

type Reaggregator struct {
	MongoClient      *mongo.Client
	DbName, CollName string
	Buckets          []ReagrBucket
}

func (a *Reaggregator) Do(ctx context.Context) error {

	for _, bk := range a.Buckets {
		err := a.process(ctx, bk)
	}

	return nil
}

func (a *Reaggregator) process(ctx context.Context, bk ReagrBucket) error {
	coll := a.MongoClient.Database(a.DbName).Collection(a.CollName)
	now := time.Now()
	cursor, err := coll.Find(ctx, bson.M{
		"type": bk.SrcType,
		"time_stamp": bson.M{
			"$lte": time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Add(-1 * bk.Age)},
	})
}
