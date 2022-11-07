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
	db                     *mongo.Client
	dbName, collName       string
	intervalForgivenessPrc float64
}

// NewAccessor returns access to db
func NewAccessor(db *mongo.Client, dbName, collName string, intervalForgivenessPrc float64) *DBAccessor {
	return &DBAccessor{db: db, dbName: dbName, collName: collName, intervalForgivenessPrc: intervalForgivenessPrc}
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

	res, err := d.everythingIsMatching(ctx, name, from, to, interval)
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return res, nil
	}

	res, err = d.aggregateSmallerInterval(ctx, name, from, to, interval)
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return res, nil
	}

	res, err = d.approximateInterval(ctx, name, from, to, interval)
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return res, nil
	}

	// cannot even approximate
	return []metric.Entry{}, nil
}

// FindAll gets all entries for the specified timeframe and interval from db
func (d *DBAccessor) FindAll(ctx context.Context, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	var results []metric.Entry
	// find all available metrics for the specified timeframe
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
		return []metric.Entry{}, nil
	}

	for _, name := range metricsList {
		res, err := d.everythingIsMatching(ctx, name, from, to, interval)
		if err != nil {
			return nil, err
		}
		if len(res) > 0 {
			results = append(results, res...)
			continue
		}

		res, err = d.aggregateSmallerInterval(ctx, name, from, to, interval)
		if err != nil {
			return nil, err
		}
		if len(res) > 0 {
			results = append(results, res...)
			continue
		}

		res, err = d.approximateInterval(ctx, name, from, to, interval)
		if err != nil {
			return nil, err
		}
		if len(res) > 0 {
			results = append(results, res...)
			continue
		}
	}
	return results, nil
}

// EverythingIsMatching finds all documents that are matching the metric, interval and timeframe
func (d *DBAccessor) everythingIsMatching(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
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

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to get a list of all returned documents for %v metric: %w", name, err)
	}

	if len(results) == 0 {
		return []metric.Entry{}, nil
	}

	return results, nil
}

// AggregateSmallerInterval aggregates all documents that are matching the metric, timeframe from a smaller interval
func (d *DBAccessor) aggregateSmallerInterval(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	var results []metric.Entry

	collection := d.db.Database(d.dbName).Collection(d.collName)

	var intervalList []time.Duration
	list, err := collection.Distinct(ctx, "type", bson.D{{"name", name}, {"type", bson.D{{"$lt", interval}}}, {"time_stamp", bson.D{{"$gte", from}, {"$lte", to}}}})

	if len(list) == 0 {
		return []metric.Entry{}, nil
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
		return []metric.Entry{}, nil
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

	//dict := make(map[string]metric.Entry)

	// aggregate available interval
	for cursor.Next(ctx) {
		var result metric.Entry
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode from db: %w", err)
		}
		results = aggrProcess(results, result, interval)
	}

	// append dict values to the final result
	//for _, v := range dict {
	//	results = append(results, v)
	//}

	return results, nil
}

// ApproximateInterval can approximate the requested interval
func (d *DBAccessor) approximateInterval(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
	var results []metric.Entry

	collection := d.db.Database(d.dbName).Collection(d.collName)

	// to find interval within 25% of requested
	lowerInterval := time.Second * time.Duration(interval.Seconds()*(1-d.intervalForgivenessPrc))
	upperInterval := time.Second * time.Duration(interval.Seconds()*(1+d.intervalForgivenessPrc))

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

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to get a list of all returned documents for %v metric: %w", name, err)
	}

	if len(results) == 0 {
		return []metric.Entry{}, nil
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
