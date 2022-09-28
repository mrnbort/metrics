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
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
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
	}
}

func TestService_deleteMetric(t *testing.T) {
	strg := &StorageMock{
		DeleteFunc: func(ctx context.Context, m metric.Entry) error {
			return nil
		},
		UpdateFunc: func(ctx context.Context, m metric.Entry) error {
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
	}
}
