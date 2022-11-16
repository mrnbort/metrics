# Holiday service [![Run Tests and Build an Image](https://github.com/mrnbort/metrics/actions/workflows/ci.yml/badge.svg)](https://github.com/mrnbort/metrics/actions/workflows/ci.yml)

## Description

The Metrics server collects data for various metrics. 
The main purpose of the server is to provide a user with the metric for monitoring events during a specified timeframe with a specified interval. 
A metric is based on what the user wants to measure.
Using a rest-like API the data is received by the server to process.
This data has the name of a metric, the value and the timestamp (provided by user).
The server aggregates data for every metric based on the age and interval provided by the admin.
Aggregated data is stored in a MongoDB. The aggregator that processes POST requests ensures that the maximum possible loss of data is no more than 1 minute of data collected. 
Aggregator is also able to clear data that already exists in the MongoDB from local memory to prevent out-of-memory errors.
Although the intervals are pre-defined, the user is able to request any frequency, and the server will calculate the metric values for the requested frequency based on the recorded data.

## Run in Docker

1. Copy docker-compose.yml

    - change the ports if needed
    - for nginx service, change `volumes` to your service config

2. Create or copy `etc/service.conf` and modify to your service config
3. Start a container with `docker-compose up`

## API

### Public Endpoints

1. `GET /get-metrics-list` - returns a list of available metrics, i.e.
    ```
   [
    "file_1",
    "file_2",
    "file_3"
   ]
    ```
2. `POST /get-metric` - returns requested metric data with a specified interval for a specified timeframe, i.e.
   - Request body:
     ```json
     {
     "name": "file_1", 
     "from": "2022-11-15T14:00:00Z", 
     "to": "2022-11-15T15:00:00Z", 
     "interval": "30m"
     }
     ```
   - Returns:
     ```json
     [
     {
     "name": "file_1",
     "time_stamp": "2022-11-15T14:30:00Z",
     "value": 1,
     "type": 1800000000000,
     "type_str": "30m0s"
     },
     {
     "name": "file_1",
     "time_stamp": "2022-11-15T15:00:00Z",
     "value": 2,
     "type": 1800000000000,
     "type_str": "30m0s"
     }
     ]
     ```

3. `POST /get-metrics` - returns all metrics data with a specified interval for a specified timeframe, i.e.
   - Request body:
      ```json
      {
      "from": "2022-11-15T14:00:00Z", 
      "to": "2022-11-15T15:00:00Z", 
      "interval": "30m"
      }
      ```
   - Returns:
     ```json
     [
     {
     "name": "file_1",
     "time_stamp": "2022-11-15T14:30:00Z",
     "value": 1,
     "type": 1800000000000,
     "type_str": "30m0s"
     },
     {
     "name": "file_1",
     "time_stamp": "2022-11-15T15:00:00Z",
     "value": 2,
     "type": 1800000000000,
     "type_str": "30m0s"
     },
     {
     "name": "file_2",
     "time_stamp": "2022-11-15T14:30:00Z",
     "value": 14,
     "type": 1800000000000,
     "type_str": "30m0s"
     },
     {
     "name": "file_2",
     "time_stamp": "2022-11-15T15:00:00Z",
     "value": 9,
     "type": 1800000000000,
     "type_str": "30m0s"
     }
     ]
     ```
  
### Protected Endpoints

1. `POST /metric` - adds a metric entry (uses Basic Auth)

    - Request body:
        ```json
        {
        "name": "file_1", 
        "time_stamp": "2022-11-15T11:04:05Z", 
        "value": 2
        }
        ```
    - Returns:
        ```json
        {
        "status": "ok" 
        }
        ```

2. `DELETE /metric?name=METRIC_NAME` - removes a metric record (uses Basic Auth)

    - Returns: 
        ```json
        {
        "status": "ok" 
        }
        ```