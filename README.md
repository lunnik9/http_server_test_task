# Multiplexer

Simple service for multiplexing queries. Fetches given URLs.

## Features

An application got two enhancements beyond standart requirements

1. Cache. An app can cache responses for later reuse. Cached value invalidates after ttl expired
2. Retryer. Until amount of simultaneous requests has not reached a certain threshold, an app can retry sending requests instead of simply returning an error. Feature was introduced because currently unused capacities could posotively affect user experience, when user can;t get their responses due to temporary internet problems, ...  

### Endpoint

#### /fetch

fetches provided urls. if any returns error, all processing would stop


`curl -X POST http://localhost:8080/fetch \
     -H "Content-Type: application/json" \
     -d '{"urls":["https://google.com","https://example.com","https://github.com","https://stackoverflow.com"]}'`

### Configs

`FETCH_TIMEOUT`

time out for every single request

`MAX_URLS`

maximum urls fetcher can possibly proces

`REQUESTS_LIMIT`

limit of requests per second

`WORKER_LIMIT`

limits of simultaneously working goroutines

`RETRY`

bool flag whether the app should retry on failed queries 

`RETRY_NUM`

how many times the app retries

`RETRY_DELAY`

delay between retries

`RETRY_FILL_RATIO`

until which point app retries instead of returning an error

`CACHE`

should an app use cache

`CACHE_TTL`

ttl for cache values

you can specify these variables w/ env variables