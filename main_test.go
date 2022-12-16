package main

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func Test_Main(t *testing.T) {

	port := 40000 + int(rand.Int31n(10000))
	os.Args = []string{"test",
		"--port=:" + strconv.Itoa(port),
		"--dbname=test",
	}

	done := make(chan struct{})
	go func() {
		<-done
		e := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		require.NoError(t, e)
	}()

	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
	}()

	// defer cleanup because require check below can fail
	defer func() {
		close(done)
		<-finished
	}()

	waitForHTTPServerStart(port)
	time.Sleep(time.Second)

	{
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", port))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "pong", string(body))
	}
	{
		client := http.Client{Timeout: 10 * time.Second}
		tm := time.Date(2022, 8, 3, 16, 23, 45, 0, time.UTC)
		resp, err := client.Post(fmt.Sprintf("http://127.0.0.1:%d/metric", port), "application/json",
			strings.NewReader(fmt.Sprintf(`{"name": "test", "value":123, "time_stamp": "%s"}`, tm.Format(time.RFC3339))))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
	{
		client := http.Client{Timeout: 10 * time.Second}
		tm := time.Date(2022, 8, 3, 16, 23, 45, 0, time.UTC)
		req, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%d/metric", port),
			strings.NewReader(fmt.Sprintf(`{"name": "test", "value":123, "time_stamp": "%s"}`, tm.Format(time.RFC3339))))
		require.NoError(t, err)
		req.SetBasicAuth("admin", "Lapatusik")
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `{"status":"ok"}`+"\n", string(body))
	}
	{
		time.Sleep(time.Minute)
		client := http.Client{Timeout: 10 * time.Second}
		tmFrom := time.Date(2022, 8, 3, 1, 23, 45, 0, time.UTC)
		tmTo := tmFrom.Add(24 * time.Hour)
		interval := time.Minute * 30
		req, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%d/get-metric", port),
			strings.NewReader(fmt.Sprintf(`{"name": "test", "from": "%s", "to": "%s", "interval": %d}`,
				tmFrom.Format(time.RFC3339), tmTo.Format(time.RFC3339), interval)))
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, `[{"name":"test","time_stamp":"2022-08-03T16:30:00Z","value":123,"type":1800000000000,"type_str":"30m0s"}]`+"\n", string(data))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	dbConn, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	err = dbConn.Database("test").Collection("metrics").Drop(ctx)
	require.NoError(t, err)
}

func waitForHTTPServerStart(port int) {
	// wait for up to 10 seconds for server to start before returning it
	client := http.Client{Timeout: time.Second}
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 100)
		if resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port)); err == nil {
			_ = resp.Body.Close()
			return
		}
	}
}
