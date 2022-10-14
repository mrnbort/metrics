package storage

import (
	"context"
	"fmt"
	"github.com/umputun/metrics/metric"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"sort"
	"time"
)

//go:generate moq -out dbaccessor_mock.go . DBAccessor

// DBAccessor initiates MongoDB
type DBAccessor struct {
	db               *mongo.Client
	dbName, collName string
}

// NewAccessor returns access to db
func NewAccessor(db *mongo.Client, dbName, collName string) *DBAccessor {
	return &DBAccessor{db: db, dbName: dbName, collName: collName}
}

// Write inserts entries to db
func (d *DBAccessor) Write(ctx context.Context, m metric.Entry) error {
	m.TimeStamp = roundUpTime(m.TimeStamp, 1*time.Minute)
	collection := d.db.Database(d.dbName).Collection(d.collName)
	m.Type = 1 * time.Minute
	m.TypeStr = "1m"
	if _, err := collection.InsertOne(ctx, m); err != nil {
		return fmt.Errorf("failed to write %+v: %w", m, err)
	}
	log.Printf("inserted metric: %v", m.Name)
	return nil
}

// Delete removes entries from db
func (d *DBAccessor) Delete(ctx context.Context, m metric.Entry) error {
	collection := d.db.Database(d.dbName).Collection(d.collName)
	if _, err := collection.DeleteMany(ctx, bson.D{{"name", m.Name}}); err != nil {
		return fmt.Errorf("failed to delete %v: %w", m.Name, err)
	}
	fmt.Printf("deleted metric %v\n", m.Name)
	return nil
}

// GetMetricsList gets a list of available metrics in db
func (d *DBAccessor) GetMetricsList(ctx context.Context) ([]string, error) {
	var metricsList []string

	collection := d.db.Database(d.dbName).Collection(d.collName)
	list, err := collection.Distinct(ctx, "name", bson.D{})
	if err != nil {
		return metricsList, fmt.Errorf("failed to read all documents: %w", err)
	}

	for _, l := range list {
		metricsList = append(metricsList, l.(string))
	}
	return metricsList, nil
}

// FindOneMetric gets the values for the required metric, timeframe and interval from db
func (d *DBAccessor) FindOneMetric(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	// 1. everything is matching
	// 2. remainder = 0: can create from existing smaller interval, and it exists in the time range
	// 3. can only create from 1 minute interval
	// 4. can approximate
	// 5. only error cannot even approximate

	res, err := d.EverythingIsMatching(ctx, name, from, to, interval)
	if err != nil {
		return nil, err
	}
	if res != nil {
		return res, nil
	}

	res, err = d.AggregateSmallerInterval(ctx, name, from, to, interval)
	if err != nil {
		return nil, err
	}
	if res != nil {
		return res, nil
	}

	res, err = d.ApproximateInterval(ctx, name, from, to, interval)
	if err != nil {
		return nil, err
	}
	if res != nil {
		return res, nil
	}

	// 5. only error cannot even approximate
	return nil, fmt.Errorf("no data for selected metric, dataframe and interval")
}

// FindAll gets all entries for the specified timeframe and interval from db
func (d *DBAccessor) FindAll(ctx context.Context, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	var results []metric.Entry
	// find all available metrics for the specified timeframe and interval
	var metricsList []string

	collection := d.db.Database(d.dbName).Collection(d.collName)
	list, err := collection.Distinct(ctx, "name", bson.M{
		"time_stamp": bson.M{
			"$gte": from,
			"$lte": to,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read all documents: %w", err)
	}

	for _, l := range list {
		metricsList = append(metricsList, l.(string))
	}

	if len(metricsList) == 0 {
		return nil, fmt.Errorf("no metrics available for requested timeframe")
	}

	for _, name := range metricsList {
		res, err := d.EverythingIsMatching(ctx, name, from, to, interval)
		if err != nil {
			return nil, err
		}
		if res != nil {
			results = append(results, res...)
			return results, nil
		}

		res, err = d.AggregateSmallerInterval(ctx, name, from, to, interval)
		if err != nil {
			return nil, err
		}
		if res != nil {
			results = append(results, res...)
			return results, nil
		}

		res, err = d.ApproximateInterval(ctx, name, from, to, interval)
		if err != nil {
			return nil, err
		}
		if res != nil {
			results = append(results, res...)
			return results, nil
		}

		// 5. only error cannot even approximate
		return nil, fmt.Errorf("no data for the dataframe and interval")
	}
	return nil, nil
}

// EverythingIsMatching finds all documents that are matching the metric, interval and timeframe
func (d *DBAccessor) EverythingIsMatching(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	var results []metric.Entry

	collection := d.db.Database(d.dbName).Collection(d.collName)

	cursor, err := collection.Find(ctx, bson.M{
		"name": name,
		"type": interval,
		"time_stamp": bson.M{
			"$gte": from,
			"$lte": to,
		},
	})
	if err != nil {
		return nil, err
	}

	if cursor.RemainingBatchLength() == 0 {
		return nil, fmt.Errorf("failed to find matching docs in db")
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to get a list of all returned documents for %v metric: %w", name, err)
	}
	return results, nil
}

// AggregateSmallerInterval aggregates all documents that are matching the metric, timeframe from a smaller interval
func (d *DBAccessor) AggregateSmallerInterval(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	var results []metric.Entry

	collection := d.db.Database(d.dbName).Collection(d.collName)

	var intervalList []time.Duration
	list, err := collection.Distinct(ctx, "type", bson.D{{"name", name}, {"type", bson.D{{"$lt", interval}}}, {"time_stamp", bson.D{{"$gte", from}, {"$lte", to}}}})

	if len(list) == 0 {
		return nil, fmt.Errorf("no metric data for this timeframe: %v - %v", from, to)
	}

	if err != nil {
		return nil, err
	}

	for _, l := range list {
		intervalList = append(intervalList, time.Duration(l.(int64)))
	}

	// sort the available intervals (descending)
	sort.Slice(intervalList, func(i, j int) bool { return intervalList[i] > intervalList[j] })

	var sInterval time.Duration

	// find the largest interval which results in 0 remainder
	for _, l := range intervalList {
		if interval%l == 0 {
			sInterval = l
			break
		}
	}

	if sInterval == 0 {
		return nil, fmt.Errorf("no interval that can be aggregated")
	}

	cursor, err := collection.Find(ctx, bson.M{
		"name": name,
		"type": sInterval,
		"time_stamp": bson.M{
			"$gte": from,
			"$lte": to,
		},
	})
	if err != nil {
		return nil, err
	}

	dict := make(map[string]metric.Entry)

	// aggregate available interval
	for cursor.Next(ctx) {
		var result metric.Entry
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode from db: %w", err)
		}
		dict = aggrProcess(dict, result, interval)
	}

	// append dict values to the final result
	for _, v := range dict {
		results = append(results, v)
	}

	//sort.Slice(results, func(i, j int) bool { return results[i].TimeStamp < results[j].TimeStamp })

	return results, nil
}

// ApproximateInterval can approximate the requested interval
func (d *DBAccessor) ApproximateInterval(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	var results []metric.Entry

	collection := d.db.Database(d.dbName).Collection(d.collName)

	// to find interval within 25% of requested
	lowerInterval := interval * 3 / 4
	upperInterval := interval * 5 / 4

	cursor, err := collection.Find(ctx, bson.M{
		"name": name,
		"type": bson.M{
			"$gte": lowerInterval,
			"$lte": upperInterval,
		},
		"time_stamp": bson.M{
			"$gte": from,
			"$lte": to,
		},
	})
	if err != nil {
		return nil, err
	}
	if cursor.RemainingBatchLength() == 0 {
		return nil, fmt.Errorf("failed to find matching docs in db")
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to get a list of all returned documents for %v metric: %w", name, err)
	}
	return results, nil
}

func roundUpTime(t time.Time, roundOn time.Duration) time.Time {
	var tr time.Time
	tr = t.Round(roundOn)

	if tr.Before(t) {
		tr = tr.Add(roundOn)
	}

	return tr
}
