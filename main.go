package main

import (
	"fmt"
	"time"
)

var TestData = []struct {
	metric    string
	timeStamp string
	value     int
}{
	{"file_1", "2022-07-17T05:39:07", 2},
	{"file_1", "2022-07-17T05:39:08", 1},
	{"file_1", "2022-07-17T05:39:09", 3},
	{"file_1", "2022-07-17T05:40:10", 3},
}

type DBData struct {
	Metric    string
	TimeStamp time.Time
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

	for _, m := range TestData {
		t, err := time.Parse(layout, m.timeStamp)
		if err != nil {
			fmt.Println(err)
		}
		t_hour := t.Hour()*60 + t.Minute()

		if _, ok := data[m.metric]; ok {
			if t_hour == data[m.metric].MinuteValue {
				data[m.metric].Value = data[m.metric].Value + m.value
			} else {
				db = append(db, DBData{
					Metric:    m.metric,
					TimeStamp: data[m.metric].TimeStamp,
					Value:     data[m.metric].Value,
				})
				data[m.metric] = &Metric{
					Value:       m.value,
					TimeStamp:   t,
					MinuteValue: t_hour,
				}
			}
		} else {
			data[m.metric] = &Metric{
				Value:       m.value,
				TimeStamp:   t,
				MinuteValue: t_hour,
			}
		}
	}

	//	data = append(
	//		data,
	//		MemoryData{"metric": m.metric, "time_value": MemoryData{{"minute": t_hour, "value": m.value}}})
	//	fmt.Println(data)
	//}

	//for i := 1; i < 1441; i++ {
	//	fmt.Println(i)
	//	if i == 1440 {
	//		i = 1
	//	}
	//}
}
