# URL Shortener

This repo contains the code and docker compose setups for a URL shortening service.

I wasn't sure if it would be considered valid to help make the application scale by fronting it with a nginx reverse proxy so I made 2 setups with docker-compose below. Both versions use the same url shortening service.

## Url Shortening Service Application

The URL shortening service is written in Go. It uses the fiber framework and exposes 2 endpoints: `/api/v1/shorten` and `/<shortened-url>`. The service uses an in-memory cache (mainly for alt-docker-compose.yml) and a backing redis cache (mainly for docker-compose.yml) to keep track of current URLs that have been shortened.
While there is 2 docker-compose setups, the application code paths do not change and neither does the configuration for them.

### Shortening a URL

Both docker-compose setups expose the system on port 8080. To shorten a URL make a POST request to `/api/v1/shorten` which takes a JSON body with a key `"url": "<url-to-shorten>"`.
The response will be in JSON with format `"<url-to-shorten": "/<shortened-path"`. The original URL will now be redirected at `localhost:8080/<shortened-path>`.

Example:
```
❯ curl -X POST -H "Content-Type: application/json" localhost:8080/api/v1/shorten -d '{"url": "https://news.ycombinator.com/item?id=27979879"}'
{"https://news.ycombinator.com/item?id=27979879":"/s7QNbSI"}
❯ curl -v localhost:8080/s7QNbSI
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8080 (#0)
> GET /s7QNbSI HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 302 Found
< Server: nginx/1.21.1
< Date: Wed, 28 Jul 2021 09:09:26 GMT
< Content-Length: 0
< Connection: keep-alive
< Location: https://news.ycombinator.com/item?id=27979879
<
* Connection #0 to host localhost left intact
```

## docker-compose.yml

```
                     docker-compose
                    |                                               |
user requests <---> | nginx <---> url shortener service <---> redis |
                    |                                               |
```

This setup uses a nginx reverse proxy to front the url shortener service. With this setup, the url shortening service can be scaled up to multiple docker containers to increase its potential performance. Redis will give the system consistency so that if a new instance comes up and its in-memory cache may not contain the shortened url it can be retrieved and then placed into the cache.
The `nginx.conf` contains a basic reverse proxy config with caching enabled for 1 day on HTTP 302s.

## alt-docker-compose.yml

```
                     docker-compose
                    |                                   |
user requests <---> | url shortener service <---> redis |
                    |                                   |
```

This setup assumes that using nginx to achieve better performance is off-limits and subsequently uses Redis as just a backing store, with an in-memory cache purely to improve performance on the single instance.

## Build System

The application is built using `Make` targets for building and running the various components.
`make compose` will run both `docker-compose build` and `docker-compose up` to run the entire setup. If you'd like to run the alternate docker compose setup simply run

```
# Runs docker-compose build and docker-compose up
$ make compose
$ docker-compose -f alt-docker-compose.yml up
```
