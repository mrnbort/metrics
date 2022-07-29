package storage

import "github.com/umputun/metrics/metric"

type DBAccessor struct {
	db []metric.Entry
}

func (d *DBAccessor) Write(m metric.Entry) error {
	d.db = append(d.db, m)
	return nil
}
