package storage

import (
	"context"
	"fmt"
	"github.com/umputun/metrics/metric"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ReaggrBucket struct {
	Interval time.Duration // 30m, 8h, 24h, 7d what interval we want to, to know what the type of the interval is after aggr
	Age      time.Duration // 24h, 7d, ...
	SrcType  time.Duration // to know what type of the interval we are looking for to aggr in db
}

type Reaggregator struct {
	MongoClient      *mongo.Client
	DbName, CollName string
	Buckets          []ReaggrBucket
}

func (a *Reaggregator) Do(ctx context.Context) error {

	for _, bk := range a.Buckets {
		if err := a.process(ctx, bk); err != nil {
			return fmt.Errorf("failed to aggregate db: %w", err)
		}
	}
	return nil
}

func (a *Reaggregator) process(ctx context.Context, bk ReaggrBucket) error {
	coll := a.MongoClient.Database(a.DbName).Collection(a.CollName)
	now := time.Now()
	cursor, err := coll.Find(ctx, bson.M{
		"type": bk.SrcType,
		"time_stamp": bson.M{
			"$lte": time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Add(-1 * bk.Age)},
	})

	if cursor.RemainingBatchLength() == 0 {
		return fmt.Errorf("failed to find matching docs in db")
	}

	if err != nil {
		return fmt.Errorf("error reading from the db: %w", err)
	}
	defer cursor.Close(ctx)

	results := make(map[string]metric.Entry)

	for cursor.Next(ctx) {
		var result metric.Entry
		if err := cursor.Decode(&result); err != nil {
			return fmt.Errorf("failed to decode from db: %w", err)
		}
		results = aggrProcess(results, result, bk.Interval)
	}

	// insert the aggregated metrics to db
	for _, v := range results {
		if _, err := coll.InsertOne(ctx, v); err != nil {
			return fmt.Errorf("failed to write %+v: %w", v, err)
		}
	}

	// delete the un-aggregated metrics from db
	_, err = coll.DeleteMany(ctx, bson.M{
		"type": bk.SrcType,
		"time_stamp": bson.M{
			"$lte": time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Add(-1 * bk.Age)},
	})
	if err != nil {
		return fmt.Errorf("failed to delete matching docs in db: %w", err)
	}

	return nil
}

func aggrProcess(results map[string]metric.Entry, result metric.Entry, interval time.Duration) map[string]metric.Entry {
	result.TimeStamp = roundUpTime(result.TimeStamp, interval)
	dictKey := result.Name + "+" + result.TimeStamp.String()
	v, ok := results[dictKey]
	if !ok {
		// metric not found
		result.Type = interval
		result.TypeStr = interval.String()
		results[dictKey] = result
		return results
	}

	// metric found
	v.Value += result.Value
	v.Type = interval
	v.TypeStr = interval.String()
	results[dictKey] = v
	return results
}
