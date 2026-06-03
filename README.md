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
