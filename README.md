# HP Tracker

A little personal health & fitness tracker.

## Status

Still sketching ideas. Still looking up almost _every_ bit of Go syntax.

## Goals

- Learn Golang
- Experiment with single-binary applications
- Experiment with building a CLI app
- Keep control of my data

## Running

Define the `STRAVA_CLIENT_ID` and `STRAVA_CLIENT_SECRET` environment variables.

```
go run .
```

## Building

```
CGO_ENABLED=1 go build
```
