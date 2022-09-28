package storage

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/umputun/metrics/metric"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	acc := NewAccessor(dbConn, "test", "metrics")

	//dbConn.Database("test").Collection("metrics").Find()

	err = acc.Write(ctx, metric.Entry{
		Name:      "file_1",
		TimeStamp: time.Date(2022, 7, 29, 12, 10, 23, 0, time.UTC),
		Value:     5,
	})

	require.NoError(t, err)

}
