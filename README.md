# MiniRedis

A lightweight, Redis‑compatible server written in Go, with a built‑in CLI client.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)

---

## Overview

MiniRedis is a Redis‑compatible in‑memory key‑value store built from scratch in Go. It supports a wide range of Redis commands, including strings, lists, sets, hashes, transactions, and optimistic locking (`WATCH`/`UNWATCH`). Persistence is handled via an Append‑Only File (AOF) with automatic rewriting.

This project is designed to be educational, lightweight, and easy to extend.

---

## Features

| Feature | Status |
|---------|--------|
| **RESP Protocol** | Full support |
| **Strings** (`SET`, `GET`, `DEL`, `EXISTS`) |
| **Expiry** (`EXPIRE`, `TTL`, `EXPIREAT`) |
| **Lists** (`LPUSH`, `RPUSH`, `LRANGE`, `LLEN`, `LPOP`, `RPOP`, `LPUSHX`, `RPUSHX`, `LINDEX`, `LSET`, `LTRIM`) |
| **Sets** (`SADD`, `SMEMBERS`, `SISMEMBER`, `SCARD`, `SREM`) |
| **Hashes** (`HSET`, `HGET`, `HGETALL`, `HDEL`, `HEXISTS`, `HLEN`) |
| **Transactions** (`MULTI`, `EXEC`, `DISCARD`) |
| **Optimistic Locking** (`WATCH`, `UNWATCH`) |
| **Keyspace** (`DBSIZE`, `KEYS`, `FLUSHALL`) |
| **AOF Persistence** (append + rewrite) |
| **Concurrency** (sharded locks) |
| **CLI Client** |
| **Test Suite** |

---

## Getting Started

### Prerequisites

- Go 1.21 or higher
- (Optional) `redis-cli` for interactive testing

### Running the Server

```bash
# Clone the repository
git clone https://github.com/GitShinobi/MiniRedis.git
cd redis

# Run the server
go run server.go
```

You'll see a prompt:

```
1:C 2026-09-13T11:07:36.639Z # oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
1:C 2026-09-13T11:07:36.639Z # Configuration loaded
```

### Running the Client

In a separate terminal:

```bash
go run client/client.go
```

You'll see a prompt:

```
1:M 2024-01-01T12:00:00.000Z # Server initialized
1:M 2024-01-01T12:00:00.000Z * Ready to accept connections
listening on localhost:8080
redis:8080> 
```

### Example Session

```
redis:8080> SET name Alice
OK

redis:8080> GET name
Alice

redis:8080> LPUSH colors red blue green
3

redis:8080> LRANGE colors 0 -1
[green blue red]

redis:8080> SADD fruits apple banana
2

redis:8080> SMEMBERS fruits
[apple banana]

redis:8080> HSET user name Alice age 30
2

redis:8080> HGETALL user
[name Alice age 30]

redis:8080> Q
```

### Client Commands

| Command | Description |
|---------|-------------|
| `CLEAR` | Clears the terminal screen |
| `Q` | Quits the client |

The client also supports quoted arguments for keys or values with spaces:

```
redis:8080> SET "my key" "hello world"
OK

redis:8080> GET "my key"
hello world
```

---

## Supported Commands

### Strings

| Command | Description |
|---------|-------------|
| `SET key value` | Set a string value |
| `GET key` | Get a string value |
| `DEL key [key ...]` | Delete one or more keys |
| `EXISTS key [key ...]` | Check if one or more keys exist |

### Expiry

| Command | Description |
|---------|-------------|
| `EXPIRE key seconds` | Set a TTL in seconds |
| `TTL key` | Get remaining TTL (`-1` = no expiry, `-2` = missing) |
| `EXPIREAT key timestamp` | Set expiry as Unix timestamp |

### Lists

| Command | Description |
|---------|-------------|
| `LPUSH key value [value ...]` | Push to head |
| `RPUSH key value [value ...]` | Push to tail |
| `LRANGE key start stop` | Get a range of elements |
| `LLEN key` | Get list length |
| `LPOP key [count]` | Pop from head |
| `RPOP key [count]` | Pop from tail |
| `LPUSHX key value [value ...]` | Push to head only if key exists |
| `RPUSHX key value [value ...]` | Push to tail only if key exists |
| `LINDEX key index` | Get element at index |
| `LSET key index value` | Set element at index |
| `LTRIM key start stop` | Trim list to range |

### Sets

| Command | Description |
|---------|-------------|
| `SADD key member [member ...]` | Add members |
| `SMEMBERS key` | Get all members |
| `SISMEMBER key member` | Check membership |
| `SCARD key` | Get set size |
| `SREM key member [member ...]` | Remove members |

### Hashes

| Command | Description |
|---------|-------------|
| `HSET key field value [field value ...]` | Set fields |
| `HGET key field` | Get a field |
| `HGETALL key` | Get all fields and values |
| `HDEL key field [field ...]` | Delete fields |
| `HEXISTS key field` | Check if field exists |
| `HLEN key` | Get number of fields |

### Transactions

| Command | Description |
|---------|-------------|
| `MULTI` | Start a transaction |
| `EXEC` | Execute queued commands |
| `DISCARD` | Discard the transaction |
| `WATCH key [key ...]` | Watch keys for optimistic locking |
| `UNWATCH` | Clear all watched keys |

### Keyspace

| Command | Description |
|---------|-------------|
| `DBSIZE` | Number of keys |
| `KEYS pattern` | List keys (pattern currently ignored — returns all) |
| `FLUSHALL` | Delete all keys |

### Utility

| Command | Description |
|---------|-------------|
| `PING` | Returns `PONG` |

---

## Project Structure

```
redis/
├── server.go          # Server entry point, handlers, AOF, concurrency
├── main_test.go       # Full test suite
├── resp/
│   └── resp.go        # RESP protocol encoding/decoding
├── client/
│   └── client.go      # Interactive CLI client
├── go.mod             # Go module definition
└── aof.log            # Append‑Only File (created at runtime)
```

---

## Running Tests

```bash
# Run all tests with verbose output
go test -v

# Run a specific test
go test -v -run TestPing

# Run with coverage
go test -cover
```

Expected output:

```
=== RUN   TestPing
--- PASS: TestPing (0.00s)
=== RUN   TestSetAndGet
--- PASS: TestSetAndGet (0.00s)
=== RUN   TestDel
--- PASS: TestDel (0.00s)
...
PASS
ok      redis   3.04s
```

The test suite covers:

- Strings (`Ping`, `SetAndGet`, `Del`, `Exists`)
- Expiry (`ExpireTTL`)
- Lists (`LPushRPushLLen`, `LRange`, `LPopRPop`, `LIndexLSetLTrim`)
- Sets (`SAddSMembers`, `SIsMemberSCardSRem`)
- Hashes (`HSetHGet`, `HGetAllHDelHExistsHLen`)
- Transactions (`MultiExec`, `Discard`)
- Watch (`Watch`, `Unwatch`)
- Keyspace (`DBSizeKeysFlushAll`)

---

## Persistence

### AOF (Append‑Only File)

- Every write command is appended to `aof.log`.
- On startup, the server **replays** the AOF to restore the dataset.
- A background goroutine **rewrites** the AOF every 24 hours to compact the log.

### Supported Data Types in AOF Rewrite

| Data type | Reconstructed with |
|-----------|-------------------|
| Strings | `SET key value` |
| Lists | `RPUSH key elem1 elem2 ...` |
| Sets | `SADD key member1 member2 ...` |
| Hashes | `HSET key field1 value1 field2 value2 ...` |
| Expiry | `EXPIREAT key timestamp` |

### Durability

- The command buffer flushes to disk when it exceeds 10 commands (see `FlushThreshold`).
- The AOF file is protected by `fileMu` to prevent concurrent writes during rewrite.

---

## Architecture

### Sharded Storage

- **16 shards** (`NumShards = 16`).
- Each shard is a `map[string]entry` with its own `sync.RWMutex`.
- Keys are distributed using `fnv32` hashing (`shardIndex`).

### Concurrency

- **Read‑only commands** use `RLock()` (e.g., `GET`, `LRANGE`, `SMEMBERS`).
- **Write commands** use `Lock()` (e.g., `SET`, `LPUSH`, `HSET`).
- **Global locks** guard the AOF file (`fileMu`), command buffer (`cmdBufMu`), and key versions (`keyVersionsMu`).

### Expiration

- **Lazy expiry:** keys are checked and deleted on access.
- **Active expiry:** a background goroutine sweeps expired keys every 100 ms.

### Transactions

- `MULTI` sets `inTransaction = true` and clears the per‑connection queue.
- Queued commands reply with `+QUEUED`.
- `EXEC` runs the queue atomically and returns an array reply.
- `DISCARD` clears the queue without executing.

### Optimistic Locking (`WATCH` / `UNWATCH`)

- Each key has a version number tracked in `keyVersions`.
- Every write to a key increments its version.
- `WATCH key ...` snapshots the current version of each key.
- `EXEC` compares snapshots to current versions; on mismatch, the transaction aborts and returns `(nil)`.
- `UNWATCH` clears the watch list without affecting the transaction state.

---

## Customization

### Configuration Constants

In `server.go`:

```go
const (
    NumShards       = 16              // Number of shards
    RewriteInterval = 24 * time.Hour  // AOF rewrite interval
    FlushThreshold  = 10              // Flush command buffer after N commands
)
```

### Changing the Port

**Server** — in `server.go`:

```go
listener, err := net.Listen("tcp", "localhost:8080")
```

**Client** — in `client/client.go`:

```go
conn, err := net.Dial("tcp", "localhost:8080")
```

### RESP Limits

In `resp/resp.go`:

```go
const (
    maxBulkPayload   = 512 * 1024 * 1024   // 512 MB
    maxArrayElements = 1_000_000           // 1M elements
)
```

---

##  Acknowledgements

- Inspired by [Redis](https://redis.io/)
- Built in Go

---
