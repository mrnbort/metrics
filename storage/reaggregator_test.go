package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umputun/metrics/metric"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"testing"
	"time"
)

func TestReaggregator_Do(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	err = acc.Write(ctx, metric.Entry{
		Name:      "file_2",
		TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
		Value:     5,
	})
	require.NoError(t, err)

	err = acc.Write(ctx, metric.Entry{
		Name:      "file_2",
		TimeStamp: time.Date(2022, 10, 11, 2, 11, 23, 0, time.UTC),
		Value:     9,
	})
	require.NoError(t, err)

	err = acc.Write(ctx, metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 10, 11, 2, 12, 23, 0, time.UTC),
		Value:     11,
	})
	require.NoError(t, err)

	err = acc.Write(ctx, metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 10, 11, 2, 13, 23, 0, time.UTC),
		Value:     11,
	})
	require.NoError(t, err)

	err = acc.Write(ctx, metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
		Value:     11,
	})
	require.NoError(t, err)

	// successful test
	reagg := &Reaggregator{
		MongoClient: dbConn,
		DbName:      "test",
		CollName:    "metrics",
		Buckets: []ReaggrBucket{
			{Interval: 3 * time.Minute, Age: 24 * time.Hour, SrcType: 1 * time.Minute},
		},
	}

	err = reagg.Do(ctx)
	require.NoError(t, err)

	var results []metric.Entry
	cursor, err := dbConn.Database("test").Collection("metrics").Find(ctx, bson.M{})
	require.NoError(t, err)

	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, 3, len(results))

	// failed test due to no data old enough
	reagg = &Reaggregator{
		MongoClient: dbConn,
		DbName:      "test",
		CollName:    "metrics",
		Buckets: []ReaggrBucket{
			{Interval: 30 * time.Minute, Age: 12000 * time.Hour, SrcType: 3 * time.Minute},
		},
	}

	err = reagg.Do(ctx)

	assert.Equal(t, nil, err)
}
