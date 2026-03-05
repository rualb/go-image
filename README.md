# go-image

`go-image` is a microservice built with Go, designed for managing, processing, and serving images. 

## Features

- **Image Processing:** Utilizes `golang.org/x/image` for image manipulation and processing tasks.
- **Storage:** Integrates with local file systems or blob storage via Docker volumes.
- **Database:** Uses GORM and PostgreSQL to store image metadata.
- **HTTP Server:** Exposes endpoints via the Echo framework.
- **Metrics:** Prometheus monitoring integration.

## Prerequisites

- Go 1.26+
- Python 3.x

## Build and Run

```sh
# Run tests
python Makefile.py test

# Build binary for Linux
python Makefile.py linux
```

## Architecture Context

In the broader system, `go-image` acts as the dedicated service for handling media. It is typically deployed alongside a reverse proxy like `go-proxy` and storage volumes for persistent image caching.
