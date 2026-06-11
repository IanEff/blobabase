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

### Store a blob — `/set`

Each query pair is stored as a key/value. Returns `204 No Content` on
success, `400 Bad Request` if no query pair is passed. Any method works, so
you can hit it from a browser address bar.

```sh
curl 'http://localhost:4000/set?greeting=hello'
```

The value is taken verbatim from the query string, so URL-encode anything with
spaces or special characters. `--data-urlencode` makes this painless:

```sh
curl -G 'http://localhost:4000/set' --data-urlencode 'greeting=hello, world!'
```

At least one pair is required; passing none returns `400`:

```sh
curl -i 'http://localhost:4000/set'
# HTTP/1.1 400 Bad Request
# a key=value query parameter is required
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
curl 'http://localhost:4000/set?foo=bar'
curl 'http://localhost:4000/get?key=foo'   # -> bar
```
