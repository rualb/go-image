# go-image

`go-image` is a high-performance, stateless image processing service written in Go. It provides on-the-fly image resizing, watermarking, and multi-tier caching. It is designed to sit behind a reverse proxy or CDN to serve optimized assets for web and mobile applications.

## Features

- **Dynamic Resizing**: Generates image variants based on configurable size steps.
- **Automated Watermarking**: Applies text-based watermarks to images exceeding specific size thresholds.
- **Multi-Bucket Support**: Organize images into separate "buckets" with unique configurations (source paths, cache paths, quality settings).
- **Directory Partitioning**: Automatically organizes source and cache files into subdirectories based on IDs to maintain filesystem performance.
- **Efficient Caching**: Processes images once and serves subsequent requests from a local cache.
- **Observability**: Built-in Prometheus metrics and health check endpoints.
- **Graceful Shutdown**: Handles OS signals to ensure requests are finished before exiting.
- **Configurable**: Setup via CLI flags, environment variables, or JSON configuration files.

## API Usage

The service exposes a simple REST API for image retrieval:

### Get Image Variant
`GET /image/api/size/:bucket/:id/:variant.jpg`

- **:bucket**: The name of the configured image bucket.
- **:id**: The unique identifier of the image (e.g., `prod-12345`).
- **:variant**: The size variant number (e.g., `1.jpg`, `2.jpg`). The actual pixel width is calculated as `variant * size_step`.

### System Endpoints
- **Health Check**: `GET /health` (Returns 200 OK)
- **Metrics**: `GET /sys/api/metrics` (Prometheus format, requires `X-Authorization` or query parameter API key).
- **Ping**: `GET /image/api/ping`

## Configuration

The application can be configured using environment variables (prefixed with `APP_`) or JSON files.

### Key Environment Variables

| Variable | Description | Default |
| :--- | :--- | :--- |
| `APP_ENV` | Environment (development, testing, production) | `production` |
| `APP_CONFIG` | Path to directory containing `config.json` | `.` |
| `APP_LISTEN` | HTTP server listen address | `127.0.0.1:32180` |
| `APP_VOLUME_DIR` | Base directory for image storage | `/app/blob` |
| `APP_IMAGE_BUCKET` | JSON array of bucket names to initialize | `[]` |
| `APP_HTTP_SYS_API_KEY` | API Key required for metrics access | (Required for metrics) |

### Bucket Configuration

Each bucket in `image_buckets` supports:
- `name`: Unique identifier.
- `source`: Path to original images.
- `cache`: Path to processed images.
- `size_step`: Pixel increment per variant (e.g., `200` means variant 1 is 200px, variant 2 is 400px).
- `watermark`: Text to overlay on the image.
- `watermark_after`: Width threshold (in px) above which the watermark is applied.

## Directory Structure

To optimize performance, `go-image` expects/creates a partitioned directory structure:
If an ID is `item-123-abc`, the service looks for:
`{source_dir}/item-123/item-123-abc.jpg`

## Deployment

### Docker

The service is designed to run in lightweight containers (e.g., Alpine Linux).

```yaml
services:
  go-image:
    image: alpine:3.20
    container_name: go-image
    command: ./go-image
    environment:
      - APP_CONFIG=/app/configs
      - APP_VOLUME_DIR=/app/data
      - APP_IMAGE_BUCKET=["products", "profiles"]
    volumes:
      - ./configs:/app/configs:ro
      - ./bin/go-image:/app/go-image
      - /mnt/storage/images:/app/data/products:ro
      - /mnt/storage/cache:/app/data/products-cache
    ports:
      - "32180:32180"
```

## Development

### Prerequisites
- Go 1.23+
- Libs for image processing (standard library handles JPEG decoding/encoding)

### Build
```bash
go build -o go-image cmd/go-image/main.go
```

### Running Tests
```bash
# Run unit tests
go test ./internal/...

# Run e2e tests
go test ./test/e2e/...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.