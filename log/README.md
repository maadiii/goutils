# Log

Zap logger wrapper with middleware support for Echo and Hertz frameworks, featuring structured logging and log rotation.

## Overview

The log package provides a production-ready logging solution built on top of Uber's Zap logger with built-in middleware for popular Go web frameworks and support for log rotation.

## Installation

```bash
go get github.com/maadiii/goutils/log
```

## Features

- 📝 Structured logging with Zap
- 🔄 Log rotation with lumberjack
- 🌐 Echo middleware integration
- ⚡ Hertz middleware integration
- 📊 Multiple output destinations
- ⚙️ Configurable log levels
- 📦 Batch logging support

## Usage

### Basic Setup

```go
package main

import (
    "github.com/maadiii/goutils/log"
    "go.uber.org/zap"
)

func main() {
    // Create logger with configuration
    logger := log.Zap(log.Config{
        Level:   "info",
        Service: "my-app",
        Env:     "production",
        Writer: log.WriterConfig{
            Stdout: true,
            FileConfig: &log.FileConfig{
                Filename:   "/var/log/app.log",
                MaxSize:    100, // MB
                MaxBackups: 3,
                MaxAge:     28,  // days
                Compress:   true,
            },
        },
    })

    // Use logger
    logger.Info("application started",
        zap.String("version", "1.0.0"),
        zap.Int("port", 8080),
    )

    logger.Error("failed to connect",
        zap.Error(err),
        zap.String("host", "localhost"),
    )
}
```

### Echo Middleware

```go
import (
    "github.com/labstack/echo/v4"
    "github.com/maadiii/goutils/log"
)

func main() {
    logger := log.Zap(log.Config{
        Level:   "info",
        Service: "api-server",
        Env:     "production",
        Writer: log.WriterConfig{
            Stdout: true,
        },
    })

    e := echo.New()

    // Add logging middleware
    e.Use(log.ZapEchoMiddleware(logger))

    e.GET("/users", func(c echo.Context) error {
        // Logs will include request details
        return c.JSON(200, users)
    })

    e.Start(":8080")
}
```

### Hertz Middleware

```go
import (
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/maadiii/goutils/log"
)

func main() {
    logger := log.Zap(log.Config{
        Level:   "info",
        Service: "api-server",
        Env:     "production",
        Writer: log.WriterConfig{
            Stdout: true,
        },
    })

    h := server.Default()

    // Add logging middleware
    h.Use(log.HertzMiddleware(logger))

    h.GET("/users", func(ctx context.Context, c *app.RequestContext) {
        // Logs will include request details
        c.JSON(200, users)
    })

    h.Spin()
}
```

## Configuration

### Config Structure

```go
type Config struct {
    Level   string       // Log level: debug, info, warn, error
    Service string       // Service name
    Env     string       // Environment: dev, staging, production
    Writer  WriterConfig // Output configuration
}

type WriterConfig struct {
    Stdout     bool        // Write to stdout
    FileConfig *FileConfig // File output configuration (optional)
}

type FileConfig struct {
    Filename   string // Log file path
    MaxSize    int    // Maximum size in megabytes
    MaxBackups int    // Maximum number of old files to keep
    MaxAge     int    // Maximum days to retain old files
    Compress   bool   // Compress rotated files
}
```

### Configuration Examples

#### Development Configuration

```go
logger := log.Zap(log.Config{
    Level:   "debug",
    Service: "my-app",
    Env:     "development",
    Writer: log.WriterConfig{
        Stdout: true,
    },
})
```

#### Production Configuration

```go
logger := log.Zap(log.Config{
    Level:   "info",
    Service: "my-app",
    Env:     "production",
    Writer: log.WriterConfig{
        Stdout: true,
        FileConfig: &log.FileConfig{
            Filename:   "/var/log/my-app.log",
            MaxSize:    100,  // 100 MB
            MaxBackups: 5,    // Keep 5 old files
            MaxAge:     30,   // 30 days
            Compress:   true, // Compress old files
        },
    },
})
```

#### Stdout Only (Containers/Cloud)

```go
logger := log.Zap(log.Config{
    Level:   "info",
    Service: "my-app",
    Env:     "production",
    Writer: log.WriterConfig{
        Stdout: true,
    },
})
```

## Logging Levels

### Available Levels

- **debug**: Detailed information for debugging
- **info**: General informational messages
- **warn**: Warning messages for potentially harmful situations
- **error**: Error messages for errors that need attention
- **fatal**: Critical errors that cause program termination
- **panic**: Severe errors that cause panic

### Usage Examples

```go
// Debug - development details
logger.Debug("processing request",
    zap.String("user_id", "123"),
    zap.Any("params", params),
)

// Info - general information
logger.Info("user logged in",
    zap.String("email", "user@example.com"),
    zap.Time("timestamp", time.Now()),
)

// Warn - potential issues
logger.Warn("slow database query",
    zap.Duration("duration", 5*time.Second),
    zap.String("query", "SELECT * FROM users"),
)

// Error - errors that occurred
logger.Error("failed to send email",
    zap.Error(err),
    zap.String("recipient", "user@example.com"),
)

// Fatal - critical errors (exits program)
logger.Fatal("database connection failed",
    zap.Error(err),
    zap.String("host", dbHost),
)
```

## Structured Logging

### Field Types

```go
import "go.uber.org/zap"

logger.Info("event occurred",
    zap.String("key", "value"),
    zap.Int("count", 42),
    zap.Int64("user_id", 123456),
    zap.Float64("price", 99.99),
    zap.Bool("success", true),
    zap.Duration("latency", 250*time.Millisecond),
    zap.Time("timestamp", time.Now()),
    zap.Error(err),
    zap.Any("data", complexStruct),
    zap.Strings("tags", []string{"tag1", "tag2"}),
)
```

### Contextual Logging

```go
// Create logger with default fields
userLogger := logger.With(
    zap.String("user_id", "123"),
    zap.String("session_id", "abc"),
)

// All logs from userLogger include these fields
userLogger.Info("action performed")
userLogger.Warn("unusual activity detected")
```

## Middleware Logging

### Echo Middleware Output

The Echo middleware automatically logs:

- HTTP method
- Request path
- Status code
- Response time
- Client IP
- User agent

Example log output:

```json
{
  "level": "info",
  "time": "2025-12-30T10:15:30Z",
  "message": "request completed",
  "service": "api-server",
  "env": "production",
  "method": "GET",
  "path": "/api/users",
  "status": 200,
  "latency": "23ms",
  "ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0..."
}
```

### Hertz Middleware Output

Similar to Echo, provides comprehensive request logging:

```json
{
  "level": "info",
  "time": "2025-12-30T10:15:30Z",
  "message": "request completed",
  "service": "api-server",
  "method": "POST",
  "path": "/api/orders",
  "status": 201,
  "latency": "145ms",
  "ip": "10.0.0.1"
}
```

## Advanced Features

### Batch Logging

The package supports batch logging for high-throughput scenarios:

```go
logger := log.Zap(log.Config{
    Level:     "info",
    Service:   "high-traffic-app",
    BatchSize: 100,           // Batch 100 logs
    BatchDur:  time.Second,   // Or flush every second
    Writer: log.WriterConfig{
        Stdout: true,
    },
})
```

### Custom Encoder

The logger uses a custom encoder with these defaults:

- Time format: RFC3339
- Level format: Lowercase
- Duration format: Seconds
- Structured JSON output

## Common Patterns

### Application Startup Logging

```go
func main() {
    logger := log.Zap(config)

    logger.Info("application starting",
        zap.String("version", version),
        zap.String("environment", env),
        zap.Int("port", port),
    )

    // ... startup code ...

    logger.Info("application ready",
        zap.Duration("startup_time", time.Since(start)),
    )
}
```

### Request Logging

```go
func handleRequest(c echo.Context) error {
    logger := c.Get("logger").(goutils.Logger)

    logger.Info("processing request",
        zap.String("endpoint", c.Path()),
        zap.String("user_id", getUserID(c)),
    )

    // ... process request ...

    logger.Info("request completed",
        zap.Int("records", len(results)),
    )

    return c.JSON(200, results)
}
```

### Error Logging with Context

```go
func processOrder(orderID string) error {
    logger := log.With(zap.String("order_id", orderID))

    logger.Info("processing order")

    if err := validateOrder(orderID); err != nil {
        logger.Error("validation failed", zap.Error(err))
        return err
    }

    if err := chargePayment(orderID); err != nil {
        logger.Error("payment failed",
            zap.Error(err),
            zap.String("payment_method", "credit_card"),
        )
        return err
    }

    logger.Info("order processed successfully")
    return nil
}
```

### Database Query Logging

```go
func queryUsers(filters UserFilters) ([]User, error) {
    start := time.Now()

    users, err := db.Query(filters)

    logger.Info("database query completed",
        zap.Duration("duration", time.Since(start)),
        zap.Int("results", len(users)),
        zap.Any("filters", filters),
        zap.Error(err),
    )

    return users, err
}
```

## Log Rotation

The package uses lumberjack for automatic log rotation:

```go
FileConfig: &log.FileConfig{
    Filename:   "/var/log/app.log",
    MaxSize:    100,  // Rotate after 100 MB
    MaxBackups: 5,    // Keep 5 old files
    MaxAge:     30,   // Delete files older than 30 days
    Compress:   true, // Compress old files with gzip
}
```

Rotated files are named: `app.log.2025-12-30T10-00-00.000.gz`

## Best Practices

1. **Use Structured Fields**

   ```go
   // Good
   logger.Info("user login", zap.String("user_id", userID))

   // Avoid
   logger.Info(fmt.Sprintf("user %s login", userID))
   ```

2. **Log Levels**

   - Use `Debug` for development details
   - Use `Info` for business events
   - Use `Warn` for recoverable issues
   - Use `Error` for failures

3. **Include Context**

   ```go
   logger.Error("operation failed",
       zap.Error(err),
       zap.String("operation", "create_user"),
       zap.String("user_id", userID),
   )
   ```

4. **Don't Log Sensitive Data**

   ```go
   // Never log passwords, tokens, credit cards
   logger.Info("user created",
       zap.String("email", email),
       // DON'T: zap.String("password", password)
   )
   ```

5. **Use Appropriate Levels in Production**

   ```go
   // Production
   Level: "info"

   // Development
   Level: "debug"
   ```

## Performance

- Zap is extremely fast (microseconds per log)
- Batch logging reduces I/O overhead
- Structured fields are more efficient than string formatting
- JSON encoding is optimized for performance

## Testing

For testing, use a no-op logger or test logger:

```go
import "go.uber.org/zap/zaptest"

func TestMyFunction(t *testing.T) {
    logger := zaptest.NewLogger(t)
    // Logs will appear in test output if test fails
}
```

## License

MIT License - see [LICENSE](../LICENSE) for details
