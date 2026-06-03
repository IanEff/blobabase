# blobabase

Blobabase is the barest of barebones KV stores.  Structure based on 'Cloud Native Go', by Matthew A. Titmus.
Yes, it initially stored blobs.

## Run

```sh
make run
```

## Build

```sh
make build
./bin/blobabase
```

## CI locally

```sh
make ci    # fmt → vet → test → build
```

## Usage (curl)

The server listens on `localhost:4000` by default (override with `-port`). All
examples assume that address.

### Store a blob — `PUT /set`

Takes `key` and `value` query params. Returns `200 OK` with an empty body on
success, `400 Bad Request` if `key` is missing.

```sh
curl -X PUT 'http://localhost:4000/set?key=greeting&value=hello'
```

The value is taken verbatim from the query string, so URL-encode anything with
spaces or special characters. `--data-urlencode` makes this painless:

```sh
curl -X PUT -G 'http://localhost:4000/set' \
  --data-urlencode 'key=greeting' \
  --data-urlencode 'value=hello, world!'
```

An empty value is allowed; only a missing `key` is rejected:

```sh
curl -i -X PUT 'http://localhost:4000/set'
# HTTP/1.1 400 Bad Request
# missing key
```

### Read a blob — `GET /get`

Takes a `key` query param. Returns `200 OK` with the stored value as the body,
or `404 Not Found` if the key was never set.

```sh
curl 'http://localhost:4000/get?key=greeting'
# hello, world!
```

Use `-i` to see the status line for a missing key:

```sh
curl -i 'http://localhost:4000/get?key=nope'
# HTTP/1.1 404 Not Found
# no such key
```

### Round-trip in one go

```sh
curl -X PUT 'http://localhost:4000/set?key=foo&value=bar'
curl 'http://localhost:4000/get?key=foo'   # -> bar
```

## Health check — `GET /healthz`

Returns `200 OK` with an empty body whenever the server is up. Handy for
liveness/readiness probes.

```sh
curl -i http://localhost:4000/healthz
# HTTP/1.1 200 OK
```

To verify health from a script, check only the status code and avoid the body:

```sh
curl -fsS -o /dev/null -w '%{http_code}\n' http://localhost:4000/healthz
# 200
```

`-f` makes curl exit non-zero on a 4xx/5xx response, so this doubles as a quick
pass/fail gate:

```sh
if curl -fsS -o /dev/null http://localhost:4000/healthz; then
  echo "blobabase is healthy"
else
  echo "blobabase is down"
fi
```
