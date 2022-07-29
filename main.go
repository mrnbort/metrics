package main

import (
	"fmt"
	"github.com/umputun/metrics/storage"
	"time"
)

var TestData = []struct {
	metric    string
	timeStamp string
	value     int
}{
	{"file_0", "2022-07-17T05:30:10", 10},
	{"file_1", "2022-07-17T05:39:07", 2},
	{"file_1", "2022-07-17T05:39:08", 1},
	{"file_1", "2022-07-17T05:39:09", 3},
	{"file_1", "2022-07-17T05:40:10", 3},
}

type DBData struct {
	Metric    string    `bson:"metric"`
	TimeStamp time.Time `bson:"-"`
	Value     int
}

type Metric struct {
	Value       int
	TimeStamp   time.Time
	MinuteValue int
}

type MemoryData map[string]*Metric

func main() {

	var db []DBData

	layout := "2006-01-02T15:04:05"

	data := make(MemoryData)

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	timeoutCh := time.After(10 * time.Minute)

	for {
		select {
		//case tick := <-ticker.C:
		case <-ticker.C:
			//tick_hour := tick.Hour()*60 + tick.Minute()
			tickHour := 340
			if len(data) > 0 {
				for k, v := range data {
					if tickHour != data[k].MinuteValue {
						db = append(db, DBData{
							Metric:    k,
							TimeStamp: v.TimeStamp,
							Value:     v.Value,
						})
						delete(data, k)
					}
				}
			}
		case <-timeoutCh:
			fmt.Println("Finished routine!")
			return
		default:
			for _, m := range TestData {
				t, err := time.Parse(layout, m.timeStamp)
				if err != nil {
					fmt.Println(err)
				}
				tHour := t.Hour()*60 + t.Minute()

				if v, ok := data[m.metric]; ok {
					if tHour == v.MinuteValue {
						data[m.metric].Value = v.Value + m.value
					} else {
						db = append(db, DBData{
							Metric:    m.metric,
							TimeStamp: v.TimeStamp,
							Value:     v.Value,
						})
						data[m.metric] = &Metric{
							Value:       m.value,
							TimeStamp:   t,
							MinuteValue: tHour,
						}
					}
				} else {
					data[m.metric] = &Metric{
						Value:       m.value,
						TimeStamp:   t,
						MinuteValue: tHour,
					}
				}
			}
			fmt.Println("done1")
		}
	}
	fmt.Println("done2")
}

func main_for_real() {
	db := &storage.DBAccessor{}

	svc := storage.New(db)
	svc.ActivateCleanup(time.Minute) // async, exit right away

	apiService := api.Service{
		svc:  svc,
		Port: port,
	}

	apiService.Run()
}
