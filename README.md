# Metrics Monitoring Service [![Run Tests and Build an Image](https://github.com/mrnbort/metrics/actions/workflows/ci.yml/badge.svg)](https://github.com/mrnbort/metrics/actions/workflows/ci.yml)

## Description

The Metrics Monitoring service acts as a rest crud service to insert update 
and retrieve metrics that allows a user to monitor metrics during a specified 
timeframe with a specified interval.

## Architecture and technical details

### Collecting data

Using a rest-like API the data is received by the server to process.
A POST request to save a metric entry contains the name of a metric, the value 
and the timestamp. This POST request uses a basic authentication method. When the request is processed by the server, the data about the 
metric is saved in the server's cache. The local memory is set up to hold only 1 
minute of data which means that if multiple POST requests are received for the same 
metric in one minute, 
they will be aggregated by summing up the values to create a one-minute interval 
value for the metric. The aggregation process in the server's cache is developed in 
order to ensure a consistent level of granularity of one minute for data that will 
be pushed to a MongoDB database. A goroutine which runs every minute was created to 
verify that as soon as the age of the metric in server's cache reaches one minute, 
the data for the metric is aggregated and pushed to the database. This two-stage 
commit logic also prevents loss of data from the local memory beyond a one-minute 
interval. 

### Data storing/management

A separate clean-up process was developed for the metrics data stored in the database.
A goroutine runs the clean-up process every 24 hours. The server admin can customize 
the criteria for which metric gets aggregated, for instance, each metric that is older
than 24 hours will be aggregated into a 5-minute interval instead of the original 
1-minute interval; each metric that is older than 7 days will be aggregated into a 
30-minute interval and so on. A DELETE request protected by a basic authentication 
method allows to delete a metric from the local memory and the database.

### Data retrieving

A user can request a list of available metrics in the MongoDB database, data for a 
specific metric for a user defined time frame and interval, and data for all the 
available metrics for a user defined time frame and interval. If the data for the metric
in the database does not match the requested interval, the server will process the 
request in one of the three ways:
- approximate the data by aggregating smaller available intervals into the requested 
interval
- approximate the data based on an admin-defined threshold, for example, if the available 
interval is within 25% of the requested interval, it will be considered as a "close match"
and provided to the user
- if no data in the database can be aggregated or "closely matched", a message "no metric
in db" will be posted

A web-based UI currently has two pages: for the list of available metrics and
for details for each of the available metrics.

### Non-functional aspects

- all the endpoints are protected against abuses with limiters
- the number of in-fly requests is also limited
- reverse-proxy in front of the running container with 
LE-based automatic SSL is set up



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
      
## command line parameters

```
Application Options:
     --port=            http data server port (default: 8080)
     --mngdburi         MongoDB uri (default: mongodb://localhost:27017)
     --dbname           MongoDB name (default: metrics-service)
     --collname         MongoDB collection name (default: metrics)
     --intforgiveprc    interval forgiveness percent which determines the acceptable deviation from the requested interval (default: 0.25)
     --cleanupdur       cleanup duration for the server's cache (default: 1m)
     --username         user name (default: admin)
     --userpasswd       user password (default: Lapatusik)
	
Help Options:
 -h, --help                Show this help message
```
