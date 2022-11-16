package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umputun/metrics/metric"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestService_postMetric(t *testing.T) {

	strg := &StorageMock{
		UpdateFunc: func(ctx context.Context, m metric.Entry) error {
			//assert.Equal(t, "test", m.Name)
			//assert.Equal(t, 123, m.Value)
			//assert.Equal(t, tm, m.TimeStamp)
			return nil
		},
		//GetFunc: func(from time.Time, to time.Time, interval time.Duration) ([]metric.Entry, error) {
		//	return []metric.Entry{{Name: "aa", Value: 123}, {}, {}}, nil
		//},
	}

	svc := &Service{Storage: strg, Auth: AuthMidlwr{
		User:   "admin",
		Passwd: "Lapatusik",
	}}

	ts := httptest.NewServer(svc.routes())
	defer ts.Close()

	client := http.Client{Timeout: time.Second}

	{ // successful attempt
		tm := time.Date(2022, 8, 3, 16, 23, 45, 0, time.UTC)
		req, err := http.NewRequest("POST", ts.URL+"/metric",
			strings.NewReader(fmt.Sprintf(`{"name": "test", "value":123, "time_stamp": "%s"}`, tm.Format(time.RFC3339))))
		require.NoError(t, err)
		req.SetBasicAuth("admin", "Lapatusik")
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"status":"ok"}`+"\n", string(data))

		require.Equal(t, 1, len(strg.UpdateCalls()))
		assert.Equal(t, "test", strg.UpdateCalls()[0].M.Name)
		assert.Equal(t, 123, strg.UpdateCalls()[0].M.Value)
		assert.Equal(t, tm, strg.UpdateCalls()[0].M.TimeStamp)
	}

	{ // failed decode
		req, err := http.NewRequest("POST", ts.URL+"/metric",
			strings.NewReader(fmt.Sprintf(``)))
		require.NoError(t, err)
		req.SetBasicAuth("admin", "Lapatusik")
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.Equal(t, 1, len(strg.UpdateCalls()))
	}

	{ // failed auth
		tm := time.Date(2022, 8, 3, 16, 23, 45, 0, time.UTC)
		req, err := http.NewRequest("POST", ts.URL+"/metric",
			strings.NewReader(fmt.Sprintf(`{"name": "test", "value":123, "time_stamp": "%s"}`, tm.Format(time.RFC3339))))
		require.NoError(t, err)
		req.SetBasicAuth("admin", "LapatusikBad")
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		require.Equal(t, 1, len(strg.UpdateCalls()))
	}

	{ // failed update
		strg.UpdateFunc = func(ctx context.Context, m metric.Entry) error {
			return errors.New("oh oh")
		}
		tm := time.Date(2022, 8, 3, 16, 23, 45, 0, time.UTC)
		req, err := http.NewRequest("POST", ts.URL+"/metric",
			strings.NewReader(fmt.Sprintf(`{"name": "test", "value":123, "time_stamp": "%s"}`, tm.Format(time.RFC3339))))
		require.NoError(t, err)
		req.SetBasicAuth("admin", "Lapatusik")
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"error":"oh oh"}`+"\n", string(data))
		require.Equal(t, 2, len(strg.UpdateCalls()))
	}
}

func TestService_deleteMetric(t *testing.T) {
	strg := &StorageMock{
		DeleteFunc: func(ctx context.Context, m metric.Entry) error {
			return nil
		},
	}
	svc := &Service{Storage: strg, Auth: AuthMidlwr{
		User:   "admin",
		Passwd: "Lapatusik",
	}}

	ts := httptest.NewServer(svc.routes())
	defer ts.Close()

	client := http.Client{Timeout: time.Second}

	{ // successful attempt
		url := fmt.Sprintf("%s/metric?name=test", ts.URL)
		req, err := http.NewRequest("DELETE", url, nil)
		require.NoError(t, err)
		req.SetBasicAuth("admin", "Lapatusik")
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"status":"ok"}`+"\n", string(data))
		require.Equal(t, 1, len(strg.DeleteCalls()))
		assert.Equal(t, "test", strg.DeleteCalls()[0].M.Name)
	}

	{ // failed auth
		url := fmt.Sprintf("%s/metric?name=test", ts.URL)
		req, err := http.NewRequest("DELETE", url, nil)
		require.NoError(t, err)
		req.SetBasicAuth("admin", "LapatusikBad")
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		require.Equal(t, 1, len(strg.DeleteCalls()))
	}

	{ // failed delete
		strg.DeleteFunc = func(ctx context.Context, m metric.Entry) error {
			return errors.New("oh oh")
		}
		req, err := http.NewRequest("DELETE", ts.URL+"/metric?name=test", nil)
		require.NoError(t, err)
		req.SetBasicAuth("admin", "Lapatusik")
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"error":"oh oh"}`+"\n", string(data))
		require.Equal(t, 2, len(strg.DeleteCalls()))
	}
}

func TestService_getMetricsList(t *testing.T) {
	strg := &StorageMock{
		GetListFunc: func(ctx context.Context) ([]string, error) {
			return []string{"file1"}, nil
		},
	}
	svc := &Service{Storage: strg}

	ts := httptest.NewServer(svc.routes())
	defer ts.Close()

	client := http.Client{Timeout: time.Second}

	{ // successful attempt
		url := fmt.Sprintf("%s/get-metrics-list", ts.URL)
		req, err := http.NewRequest("GET", url, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `["file1"]`+"\n", string(data))
		require.Equal(t, 1, len(strg.GetListCalls()))
	}

	{ // failed get list
		strg.GetListFunc = func(ctx context.Context) ([]string, error) {
			return nil, errors.New("oh oh")
		}
		req, err := http.NewRequest("GET", ts.URL+"/get-metrics-list", nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"error":"oh oh"}`+"\n", string(data))
		require.Equal(t, 2, len(strg.GetListCalls()))
	}

	{ // successful attempt but list is empty
		strg.GetListFunc = func(ctx context.Context) ([]string, error) {
			return []string{}, errors.New("no metrics in db")
		}
		req, err := http.NewRequest("GET", ts.URL+"/get-metrics-list", nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"error":"no metrics in db"}`+"\n", string(data))
		require.Equal(t, 3, len(strg.GetListCalls()))
	}
}

func TestService_getMetric(t *testing.T) {
	strg := &StorageMock{
		GetOneMetricFunc: func(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
			return []metric.Entry{
				{
					Name:      "file_1",
					TimeStamp: time.Date(2022, 10, 11, 2, 21, 23, 0, time.UTC),
					Value:     1,
				},
			}, nil
		},
	}
	svc := &Service{Storage: strg}

	ts := httptest.NewServer(svc.routes())
	defer ts.Close()

	client := http.Client{Timeout: time.Second}

	{ // successful attempt
		tmFrom := "2022-08-03T16:23:45Z"
		//tmFrom := time.Date(2022, 8, 3, 16, 23, 45, 0, time.UTC)
		tmTo := "2022-08-04T17:24:45Z"
		//tmTo := tmFrom.Add(24 * time.Hour)
		interval := time.Minute * 30
		//req, err := http.NewRequest("POST", ts.URL+"/get-metric",
		//	strings.NewReader(fmt.Sprintf(`{"name": "test", "from": "%s", "to": "%s", "interval": %d}`,
		//		tmFrom.Format(time.RFC3339), tmTo.Format(time.RFC3339), interval)))
		req, err := http.NewRequest("POST", ts.URL+"/get-metric",
			strings.NewReader(fmt.Sprintf(`{"name": "test", "from": %q, "to": %q, "interval": %d}`,
				tmFrom, tmTo, interval)))
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `[{"name":"file_1","time_stamp":"2022-10-11T02:21:23Z","value":1,"type":0,"type_str":""}]`+"\n", string(data))
		require.Equal(t, 1, len(strg.GetOneMetricCalls()))
		assert.Equal(t, time.Minute*30, strg.GetOneMetricCalls()[0].Interval)
		assert.Equal(t, "test", strg.GetOneMetricCalls()[0].Name)
		assert.Equal(t, time.Date(2022, time.August, 3, 16, 23, 45, 0, time.UTC), strg.GetOneMetricCalls()[0].From)
		assert.Equal(t, time.Date(2022, time.August, 4, 17, 24, 45, 0, time.UTC), strg.GetOneMetricCalls()[0].To)

	}

	{ // failed decode
		req, err := http.NewRequest("POST", ts.URL+"/get-metric",
			strings.NewReader(fmt.Sprintf(``)))
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.Equal(t, 1, len(strg.GetOneMetricCalls()))
	}

	{ // failed to get metric data
		strg.GetOneMetricFunc = func(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
			return nil, errors.New("oh oh")
		}
		tmFrom := time.Date(2022, 8, 3, 16, 23, 45, 0, time.UTC)
		tmTo := tmFrom.Add(24 * time.Hour)
		interval := time.Minute * 30
		req, err := http.NewRequest("POST", ts.URL+"/get-metric",
			strings.NewReader(fmt.Sprintf(`{"name": "test", "from": "%s", "to": "%s", "interval": %d}`,
				tmFrom.Format(time.RFC3339), tmTo.Format(time.RFC3339), interval)))
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"error":"oh oh"}`+"\n", string(data))
		require.Equal(t, 2, len(strg.GetOneMetricCalls()))
	}
}

func TestService_getMetrics(t *testing.T) {
	strg := &StorageMock{
		GetAllFunc: func(ctx context.Context, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
			return []metric.Entry{
				{
					Name:      "file_1",
					TimeStamp: time.Date(2022, 10, 11, 2, 21, 23, 0, time.UTC),
					Value:     1,
				},
			}, nil
		},
	}
	svc := &Service{Storage: strg}

	ts := httptest.NewServer(svc.routes())
	defer ts.Close()

	client := http.Client{Timeout: time.Second}

	{ // successful attempt
		tmFrom := time.Date(2022, 8, 3, 16, 23, 45, 0, time.UTC)
		tmTo := tmFrom.Add(24 * time.Hour)
		interval := "30m"
		req, err := http.NewRequest("POST", ts.URL+"/get-metrics",
			strings.NewReader(fmt.Sprintf(`{"from": "%s", "to": "%s", "interval": "%s"}`,
				tmFrom.Format(time.RFC3339), tmTo.Format(time.RFC3339), interval)))
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `[{"name":"file_1","time_stamp":"2022-10-11T02:21:23Z","value":1,"type":0,"type_str":""}]`+"\n", string(data))
		require.Equal(t, 1, len(strg.GetAllCalls()))
		assert.Equal(t, time.Minute*30, strg.GetAllCalls()[0].Interval)
		assert.Equal(t, time.Date(2022, 8, 3, 16, 23, 45, 0, time.UTC), strg.GetAllCalls()[0].From)
		assert.Equal(t, time.Date(2022, 8, 4, 16, 23, 45, 0, time.UTC), strg.GetAllCalls()[0].To)

	}

	{ // failed decode
		req, err := http.NewRequest("POST", ts.URL+"/get-metrics",
			strings.NewReader(fmt.Sprintf(``)))
		require.NoError(t, err)
		resp, err := client.Do(req)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		require.Equal(t, 1, len(strg.GetAllCalls()))
	}

	{ // failed to get metric data
		strg.GetAllFunc = func(ctx context.Context, from, to time.Time, interval time.Duration) ([]metric.Entry, error) {
			return nil, errors.New("oh oh")
		}
		tmFrom := time.Date(2022, 8, 3, 16, 23, 45, 0, time.UTC)
		tmTo := tmFrom.Add(24 * time.Hour)
		interval := "30m"
		req, err := http.NewRequest("POST", ts.URL+"/get-metrics",
			strings.NewReader(fmt.Sprintf(`{"from": "%s", "to": "%s", "interval": "%s"}`,
				tmFrom.Format(time.RFC3339), tmTo.Format(time.RFC3339), interval)))
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `{"error":"oh oh"}`+"\n", string(data))
		require.Equal(t, 2, len(strg.GetAllCalls()))
	}
}

func TestService_Run(t *testing.T) {
	done := make(chan struct{})
	go func() {
		<-done
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.NoError(t, e)
	}()
}
