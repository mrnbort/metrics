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
	"strconv"
	"testing"
	"time"
)

func TestDBAccessor_Write(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc, metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 55, 0, time.UTC),
		Value:     5,
	},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
			Value:     9,
		})

	var results []metric.Entry
	cursor, err := dbConn.Database("test").Collection("metrics").Find(ctx, bson.M{})
	require.NoError(t, err)

	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}
	i := 0
	for _, result := range results {
		i += result.Value
	}
	assert.Equal(t, 14, i)
}

func TestDBAccessor_Delete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc, metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     5,
	},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
			Value:     9,
		})

	err = acc.Delete(ctx, metric.Entry{
		Name: "file_2",
	})
	require.NoError(t, err)

	var results []metric.Entry
	cursor, err := dbConn.Database("test").Collection("metrics").Find(ctx, bson.M{})
	require.NoError(t, err)

	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, 1, len(results))
}

func TestDBAccessor_GetMetricsList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc, metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     5,
	},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
			Value:     9,
		},
		metric.Entry{
			Name:      "file_3",
			TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
			Value:     11,
		})

	metricsList, err := acc.GetMetricsList(ctx)
	assert.Equal(t, []string{"file_1", "file_2", "file_3"}, metricsList)
}

func TestDBAccessor_everythingIsMatching_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc, metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
		Value:     5,
	},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 11, 23, 0, time.UTC),
			Value:     9,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 12, 23, 0, time.UTC),
			Value:     11,
		})

	metricsList, err := acc.everythingIsMatching(ctx,
		"file_1",
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		1*time.Minute)
	assert.Equal(t, 3, len(metricsList))
}

// test for empty slice: nothing is matching in db
func TestDBAccessor_everythingIsMatching_None(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	metricsList, err := acc.everythingIsMatching(ctx,
		"file_1",
		time.Date(2022, 11, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 11, 11, 3, 0, 0, 0, time.UTC),
		1*time.Minute)
	assert.Equal(t, 0, len(metricsList))
}

func TestDBAccessor_aggregateSmallerInterval_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 11, 23, 0, time.UTC),
			Value:     9,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 12, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 17, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		})

	res, err := acc.aggregateSmallerInterval(ctx,
		"file_1",
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		5*time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(res))
}

// test for empty slice due to no data in the requested timeframe
func TestDBAccessor_aggregateSmallerInterval_NoData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 11, 23, 0, time.UTC),
			Value:     9,
		})

	res, err := acc.aggregateSmallerInterval(ctx,
		"file_1",
		time.Date(2022, 11, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 11, 11, 3, 0, 0, 0, time.UTC),
		5*time.Minute)
	assert.Equal(t, 0, len(res))
}

// test for empty slice due to no available interval that would result in 0 remainder
func TestDBAccessor_aggregateSmallerInterval_NoZeroRemainder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 11, 23, 0, time.UTC),
			Value:     9,
		})

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

	res, err := acc.aggregateSmallerInterval(ctx,
		"file_1",
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		5*time.Minute)

	assert.Equal(t, 0, len(res))
}

// successful test to make sure other intervals are left not aggregated
func TestDBAccessor_aggregateSmallerInterval_IntervalCheck(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 11, 23, 0, time.UTC),
			Value:     9,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 12, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 17, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		})

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

	err = acc.Write(ctx, metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
		Value:     11,
	})
	require.NoError(t, err)

	res, err := acc.aggregateSmallerInterval(ctx,
		"file_1",
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		15*time.Minute)

	require.NoError(t, err)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, 47, res[0].Value+res[1].Value)
}

func TestDBAccessor_ApproximateInterval_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 11, 23, 0, time.UTC),
			Value:     9,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 12, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 17, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		},
	)

	reagg := &Reaggregator{
		MongoClient: dbConn,
		DbName:      "test",
		CollName:    "metrics",
		Buckets: []ReaggrBucket{
			{Interval: 5 * time.Minute, Age: 24 * time.Hour, SrcType: 1 * time.Minute},
		},
	}
	err = reagg.Do(ctx)
	require.NoError(t, err)

	res, err := acc.approximateInterval(ctx,
		"file_1",
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		6*time.Minute)

	require.NoError(t, err)
	assert.Equal(t, 3, len(res))
}

// failed test due to no interval that would be within 25%
func TestDBAccessor_ApproximateInterval_NoForgiveness(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	res, err := acc.approximateInterval(ctx,
		"file_1",
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		10*time.Minute)

	assert.Equal(t, 0, len(res))
}

// failed test due to no data within the required date range
func TestDBAccessor_ApproximateInterval_NoData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	res, err := acc.approximateInterval(ctx,
		"file_1",
		time.Date(2022, 11, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 11, 11, 3, 0, 0, 0, time.UTC),
		6*time.Minute)

	assert.Equal(t, 0, len(res))
}

func Test_roundUpTime(t *testing.T) {
	tbl := []struct {
		tm      time.Time
		roundOn time.Duration
		res     time.Time
	}{
		{
			time.Date(2022, 10, 27, 0, 1, 23, 0, time.UTC),
			3 * time.Minute,
			time.Date(2022, 10, 27, 0, 3, 0, 0, time.UTC),
		},
		{
			time.Date(2022, 10, 27, 16, 23, 23, 0, time.UTC),
			15 * time.Minute,
			time.Date(2022, 10, 27, 16, 30, 0, 0, time.UTC),
		},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			assert.Equal(t, tt.res, roundUpTime(tt.tm, tt.roundOn))
		})
	}
}

// successful test when there is a matching interval
func TestDBAccessor_FindOneMetric_EverythingIsMatching(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 11, 23, 0, time.UTC),
			Value:     9,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 12, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 17, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		},
	)

	reagg := &Reaggregator{
		MongoClient: dbConn,
		DbName:      "test",
		CollName:    "metrics",
		Buckets: []ReaggrBucket{
			{Interval: 5 * time.Minute, Age: 24 * time.Hour, SrcType: 1 * time.Minute},
		},
	}
	err = reagg.Do(ctx)
	require.NoError(t, err)

	err = acc.Write(ctx, metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 10, 11, 2, 27, 23, 0, time.UTC),
		Value:     41,
	})
	require.NoError(t, err)

	err = acc.Write(ctx, metric.Entry{
		Name:      "file_2",
		TimeStamp: time.Date(2022, 10, 11, 2, 26, 23, 0, time.UTC),
		Value:     31,
	})
	require.NoError(t, err)

	res, err := acc.FindOneMetric(
		ctx,
		"file_1",
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		5*time.Minute)
	require.NoError(t, err)

	assert.Equal(t, 3, len(res))
}

// successful test when aggregating smaller interval
func TestDBAccessor_FindOneMetric_AggrSmallerInterval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 17, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 27, 23, 0, time.UTC),
			Value:     31,
		},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		},
	)

	res, err := acc.FindOneMetric(
		ctx,
		"file_1",
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		10*time.Minute)
	require.NoError(t, err)

	assert.Equal(t, 2, len(res))
	assert.Equal(t, 47, res[0].Value+res[1].Value)
}

// successful test when approximating interval
func TestDBAccessor_FindOneMetric_ApproximateInterval(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 2)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 17, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 27, 23, 0, time.UTC),
			Value:     31,
		},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		},
	)

	res, err := acc.FindOneMetric(
		ctx,
		"file_1",
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		2*time.Minute)
	require.NoError(t, err)

	assert.Equal(t, 3, len(res))
}

// test for when cannot approximate
func TestDBAccessor_FindOneMetric_CannotApprox(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 17, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 27, 23, 0, time.UTC),
			Value:     31,
		},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		},
	)

	reagg := &Reaggregator{
		MongoClient: dbConn,
		DbName:      "test",
		CollName:    "metrics",
		Buckets: []ReaggrBucket{
			{Interval: 5 * time.Minute, Age: 24 * time.Hour, SrcType: 1 * time.Minute},
		},
	}
	err = reagg.Do(ctx)
	require.NoError(t, err)

	res, err := acc.FindOneMetric(
		ctx,
		"file_1",
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		2*time.Minute)

	assert.Equal(t, 0, len(res))
}

func testWriteMany(t *testing.T, acc *DBAccessor, metrics ...metric.Entry) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, m := range metrics {
		err := acc.Write(ctx, m)
		require.NoError(t, err)
	}
}

func TestDBAccessor_FindAll_IntervMatchOrAggr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 17, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 27, 23, 0, time.UTC),
			Value:     31,
		},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 10, 11, 2, 26, 23, 0, time.UTC),
			Value:     1,
		},
		metric.Entry{
			Name:      "file_3",
			TimeStamp: time.Date(2022, 11, 11, 2, 26, 23, 0, time.UTC),
			Value:     1,
		},
	)

	res, err := acc.FindAll(
		ctx,
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		2*time.Minute)
	require.NoError(t, err)

	assert.Equal(t, 5, len(res))
}

func TestDBAccessor_FindAll_IntervApprox(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 17, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 19, 23, 0, time.UTC),
			Value:     31,
		},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 10, 11, 2, 21, 23, 0, time.UTC),
			Value:     1,
		},
		metric.Entry{
			Name:      "file_3",
			TimeStamp: time.Date(2022, 11, 11, 2, 26, 23, 0, time.UTC),
			Value:     1,
		},
	)

	reagg := &Reaggregator{
		MongoClient: dbConn,
		DbName:      "test",
		CollName:    "metrics",
		Buckets: []ReaggrBucket{
			{Interval: 5 * time.Minute, Age: 24 * time.Hour, SrcType: 1 * time.Minute},
		},
	}
	err = reagg.Do(ctx)
	require.NoError(t, err)

	res, err := acc.FindAll(
		ctx,
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		6*time.Minute)
	require.NoError(t, err)

	assert.Equal(t, 3, len(res))
}

func TestDBAccessor_FindAll_CannotApprox(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	defer func() {
		err := dbConn.Database("test").Collection("metrics").Drop(ctx)
		require.NoError(t, err)
	}()

	acc := NewAccessor(dbConn, "test", "metrics", 0.25)

	testWriteMany(t, acc,
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 10, 23, 0, time.UTC),
			Value:     5,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 17, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_1",
			TimeStamp: time.Date(2022, 10, 11, 2, 19, 23, 0, time.UTC),
			Value:     31,
		},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 10, 11, 2, 20, 23, 0, time.UTC),
			Value:     11,
		},
		metric.Entry{
			Name:      "file_2",
			TimeStamp: time.Date(2022, 10, 11, 2, 21, 23, 0, time.UTC),
			Value:     1,
		},
		metric.Entry{
			Name:      "file_3",
			TimeStamp: time.Date(2022, 11, 11, 2, 26, 23, 0, time.UTC),
			Value:     1,
		},
	)

	reagg := &Reaggregator{
		MongoClient: dbConn,
		DbName:      "test",
		CollName:    "metrics",
		Buckets: []ReaggrBucket{
			{Interval: 5 * time.Minute, Age: 24 * time.Hour, SrcType: 1 * time.Minute},
		},
	}
	err = reagg.Do(ctx)
	require.NoError(t, err)

	res, err := acc.FindAll(
		ctx,
		time.Date(2022, 10, 11, 2, 0, 0, 0, time.UTC),
		time.Date(2022, 10, 11, 3, 0, 0, 0, time.UTC),
		3*time.Minute)
	require.NoError(t, err)

	assert.Equal(t, 0, len(res))
}
