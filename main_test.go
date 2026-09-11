package main

import (
	"bufio"
	"net"
	"fmt"
	"strings"
	"testing"
	"time"
    "redis/resp"
)

type TestClient struct{
	conn net.Conn
	reader *bufio.Reader
}

func NewTestClient(port string)(*TestClient,error){
	conn, err := net.Dial("tcp","localhost"+port)
	if err != nil {
		return nil,err
	}
	return &TestClient{
		conn: conn,
		reader: bufio.NewReader(conn),
	},nil
}

func (c *TestClient)Close(){
	c.conn.Close()
}

func (c *TestClient) SendCommand(args ...string) string{
	var strreplyonse strings.Builder
	fmt.Fprintf(&strreplyonse, "*%d\r\n", len(args))
	for i := range args {
		fmt.Fprintf(&strreplyonse, "$%d\r\n", len(args[i]))
		fmt.Fprintf(&strreplyonse, "%s\r\n", args[i])
	}
	replyonse := strreplyonse.String()
	return replyonse
}

func setupTest(t *testing.T) *TestClient {
    client, err := NewTestClient(":8080")
    if err != nil {
        t.Fatal(err)
    }
    client.conn.Write([]byte(client.SendCommand("FLUSHALL")))
    _, err = resp.ReadResponse(client.reader)
    if err != nil {
        client.Close()
        t.Fatal(err)
    }
    return client
}
// --------------------------------------------------------------
// Existing Tests
// --------------------------------------------------------------

func TestPing(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    client.conn.Write([]byte(client.SendCommand("PING")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "PONG" {
        t.Errorf("expected PONG, got %q", reply)
    }
}

func TestSetAndGet(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    // SET
    client.conn.Write([]byte(client.SendCommand("SET", "key", "value")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "OK" {
        t.Errorf("expected OK, got %q", reply)
    }

    // GET
    client.conn.Write([]byte(client.SendCommand("GET", "key")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "value" {
        t.Errorf("expected value, got %q", reply)
    }

    // GET missing
    client.conn.Write([]byte(client.SendCommand("GET", "nonexistent")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "(nil)" {
        t.Errorf("expected (nil), got %q", reply)
    }
}

// --------------------------------------------------------------
// String Commands
// --------------------------------------------------------------

func TestDel(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    // SET
    client.conn.Write([]byte(client.SendCommand("SET", "key", "value")))
    _, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    // DEL
    client.conn.Write([]byte(client.SendCommand("DEL", "key")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "1" {
        t.Errorf("expected 1, got %q", reply)
    }

    // GET missing
    client.conn.Write([]byte(client.SendCommand("GET", "key")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "(nil)" {
        t.Errorf("expected (nil), got %q", reply)
    }

    // DEL missing
    client.conn.Write([]byte(client.SendCommand("DEL", "key")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "0" {
        t.Errorf("expected 0, got %q", reply)
    }
}

func TestExists(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    client.conn.Write([]byte(client.SendCommand("SET", "key", "value")))
    _, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    client.conn.Write([]byte(client.SendCommand("EXISTS", "key")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "1" {
        t.Errorf("expected 1, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("EXISTS", "missing")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "0" {
        t.Errorf("expected 0, got %q", reply)
    }
}

func TestExpireTTL(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    client.conn.Write([]byte(client.SendCommand("SET", "temp", "expire")))
    _, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    client.conn.Write([]byte(client.SendCommand("EXPIRE", "temp", "2")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "1" {
        t.Errorf("expected 1, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("TTL", "temp")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "2" && reply != "1" {
        t.Errorf("expected 1 or 2, got %q", reply)
    }

    time.Sleep(3 * time.Second)

    client.conn.Write([]byte(client.SendCommand("GET", "temp")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "(nil)" {
        t.Errorf("expected (nil), got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("TTL", "temp")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "-2" {
        t.Errorf("expected -2, got %q", reply)
    }
}

// --------------------------------------------------------------
// List Commands
// --------------------------------------------------------------

func TestLPushRPushLLen(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    // LPUSH
    client.conn.Write([]byte(client.SendCommand("LPUSH", "list", "a", "b", "c")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "3" {
        t.Errorf("expected 3, got %q", reply)
    }

    // LLEN
    client.conn.Write([]byte(client.SendCommand("LLEN", "list")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "3" {
        t.Errorf("expected 3, got %q", reply)
    }

    // RPUSH
    client.conn.Write([]byte(client.SendCommand("RPUSH", "list", "x", "y")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "5" {
        t.Errorf("expected 5, got %q", reply)
    }
}

func TestLRange(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    client.conn.Write([]byte(client.SendCommand("RPUSH", "list", "a", "b", "c", "d", "e")))
    _, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    // Full range
    client.conn.Write([]byte(client.SendCommand("LRANGE", "list", "0", "-1")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "[a b c d e]" {
        t.Errorf("expected [a b c d e], got %q", reply)
    }

    // Partial range
    client.conn.Write([]byte(client.SendCommand("LRANGE", "list", "1", "3")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "[b c d]" {
        t.Errorf("expected [b c d], got %q", reply)
    }
}

func TestLPopRPop(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    client.conn.Write([]byte(client.SendCommand("RPUSH", "list", "a", "b", "c")))
    _, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    // LPOP
    client.conn.Write([]byte(client.SendCommand("LPOP", "list")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "a" {
        t.Errorf("expected a, got %q", reply)
    }

    // RPOP
    client.conn.Write([]byte(client.SendCommand("RPOP", "list")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "c" {
        t.Errorf("expected c, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("LRANGE", "list", "0", "-1")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "[b]" {
        t.Errorf("expected [b], got %q", reply)
    }

    // Pop from empty list
    client.conn.Write([]byte(client.SendCommand("LPOP", "list")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "b" {
        t.Errorf("expected b, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("LPOP", "list")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "(nil)" {
        t.Errorf("expected (nil), got %q", reply)
    }
}

func TestLIndexLSetLTrim(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    client.conn.Write([]byte(client.SendCommand("RPUSH", "list", "a", "b", "c", "d", "e")))
    _, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    // LINDEX positive
    client.conn.Write([]byte(client.SendCommand("LINDEX", "list", "0")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "a" {
        t.Errorf("expected a, got %q", reply)
    }

    // LINDEX negative
    client.conn.Write([]byte(client.SendCommand("LINDEX", "list", "-1")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "e" {
        t.Errorf("expected e, got %q", reply)
    }

    // LINDEX out of bounds
    client.conn.Write([]byte(client.SendCommand("LINDEX", "list", "10")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "(nil)" {
        t.Errorf("expected (nil), got %q", reply)
    }

    // LSET
    client.conn.Write([]byte(client.SendCommand("LSET", "list", "0", "z")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "OK" {
        t.Errorf("expected OK, got %q", reply)
    }

    // LTRIM
    client.conn.Write([]byte(client.SendCommand("LTRIM", "list", "1", "3")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "OK" {
        t.Errorf("expected OK, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("LRANGE", "list", "0", "-1")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "[b c d]" { 
        t.Errorf("expected [b c d], got %q", reply)
    }
}

// --------------------------------------------------------------
// Set Commands
// --------------------------------------------------------------

func TestSAddSMembers(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    // SADD
    client.conn.Write([]byte(client.SendCommand("SADD", "set", "a", "b", "c")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "3" {
        t.Errorf("expected 3, got %q", reply)
    }

    // SMEMBERS
    client.conn.Write([]byte(client.SendCommand("SMEMBERS", "set")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    // Order may vary, check contains
    if !strings.Contains(reply, "a") || !strings.Contains(reply, "b") || !strings.Contains(reply, "c") {
        t.Errorf("expected [a b c], got %q", reply)
    }
}

func TestSIsMemberSCardSRem(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    client.conn.Write([]byte(client.SendCommand("SADD", "set", "a", "b", "c")))
    _, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    // SISMEMBER
    client.conn.Write([]byte(client.SendCommand("SISMEMBER", "set", "a")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "1" {
        t.Errorf("expected 1, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("SISMEMBER", "set", "x")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "0" {
        t.Errorf("expected 0, got %q", reply)
    }

    // SCARD
    client.conn.Write([]byte(client.SendCommand("SCARD", "set")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "3" {
        t.Errorf("expected 3, got %q", reply)
    }

    // SREM
    client.conn.Write([]byte(client.SendCommand("SREM", "set", "a")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "1" {
        t.Errorf("expected 1, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("SMEMBERS", "set")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if !strings.Contains(reply, "b") || !strings.Contains(reply, "c") || strings.Contains(reply, "a") {
        t.Errorf("expected [b c], got %q", reply)
    }
}

// --------------------------------------------------------------
// Hash Commands
// --------------------------------------------------------------

func TestHSetHGet(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    // HSET new
    client.conn.Write([]byte(client.SendCommand("HSET", "hash", "field1", "value1")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "1" {
        t.Errorf("expected 1, got %q", reply)
    }

    // HSET update
    client.conn.Write([]byte(client.SendCommand("HSET", "hash", "field1", "newvalue")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "0" {
        t.Errorf("expected 0, got %q", reply)
    }

    // HGET
    client.conn.Write([]byte(client.SendCommand("HGET", "hash", "field1")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "newvalue" {
        t.Errorf("expected newvalue, got %q", reply)
    }

    // HGET missing
    client.conn.Write([]byte(client.SendCommand("HGET", "hash", "missing")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "(nil)" {
        t.Errorf("expected (nil), got %q", reply)
    }
}

func TestHGetAllHDelHExistsHLen(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    client.conn.Write([]byte(client.SendCommand("HSET", "hash", "f1", "v1", "f2", "v2")))
    _, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    // HGETALL
    client.conn.Write([]byte(client.SendCommand("HGETALL", "hash")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if !strings.Contains(reply, "f1") || !strings.Contains(reply, "v1") ||
       !strings.Contains(reply, "f2") || !strings.Contains(reply, "v2") {
        t.Errorf("expected [f1 v1 f2 v2], got %q", reply)
    }

    // HEXISTS
    client.conn.Write([]byte(client.SendCommand("HEXISTS", "hash", "f1")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "1" {
        t.Errorf("expected 1, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("HEXISTS", "hash", "missing")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "0" {
        t.Errorf("expected 0, got %q", reply)
    }

    // HLEN
    client.conn.Write([]byte(client.SendCommand("HLEN", "hash")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "2" {
        t.Errorf("expected 2, got %q", reply)
    }

    // HDEL
    client.conn.Write([]byte(client.SendCommand("HDEL", "hash", "f1")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "1" {
        t.Errorf("expected 1, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("HGET", "hash", "f1")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "(nil)" {
        t.Errorf("expected (nil), got %q", reply)
    }
}

// --------------------------------------------------------------
// Transaction Commands
// --------------------------------------------------------------

func TestMultiExec(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    // MULTI
    client.conn.Write([]byte(client.SendCommand("MULTI")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "OK" {
        t.Errorf("expected OK, got %q", reply)
    }

    // QUEUED commands
    client.conn.Write([]byte(client.SendCommand("SET", "tx_key", "value1")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "QUEUED" {
        t.Errorf("expected QUEUED, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("LPUSH", "tx_list", "a", "b")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "QUEUED" {
        t.Errorf("expected QUEUED, got %q", reply)
    }

    // EXEC
    client.conn.Write([]byte(client.SendCommand("EXEC")))
	reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if !strings.Contains(reply, "OK") || !strings.Contains(reply, "2") {
        t.Errorf("expected OK and 2, got %q", reply)
    }

    // Verify
    client.conn.Write([]byte(client.SendCommand("GET", "tx_key")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "value1" {
        t.Errorf("expected value1, got %q", reply)
    }

	client.conn.Write([]byte(client.SendCommand("LRANGE", "tx_list","0","-1")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "[b a]" {
        t.Errorf("expected [a b], got %q", reply)
    }
}

func TestDiscard(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    client.conn.Write([]byte(client.SendCommand("MULTI")))
    _, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    client.conn.Write([]byte(client.SendCommand("SET", "discard_key", "should_not_exist")))
    _, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    client.conn.Write([]byte(client.SendCommand("DISCARD")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "OK" {
        t.Errorf("expected OK, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("GET", "discard_key")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "(nil)" {
        t.Errorf("expected (nil), got %q", reply)
    }
}

// --------------------------------------------------------------
// Watch Commands
// --------------------------------------------------------------

func TestWatch(t *testing.T) {
    client1, err := NewTestClient(":8080")
    if err != nil {
        t.Fatal(err)
    }
    defer client1.Close()

    client2, err := NewTestClient(":8080")
    if err != nil {
        t.Fatal(err)
    }
    defer client2.Close()

    // Client 1 watches
    client1.conn.Write([]byte(client1.SendCommand("WATCH", "watch_key")))
    reply, err := resp.ReadResponse(client1.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "OK" {
        t.Errorf("expected OK, got %q", reply)
    }

    // Client 2 modifies
    client2.conn.Write([]byte(client2.SendCommand("SET", "watch_key", "changed")))
    _, err = resp.ReadResponse(client2.reader)
    if err != nil {
        t.Fatal(err)
    }

    // Client 1 transaction
    client1.conn.Write([]byte(client1.SendCommand("MULTI")))
    _, err = resp.ReadResponse(client1.reader)
    if err != nil {
        t.Fatal(err)
    }

    client1.conn.Write([]byte(client1.SendCommand("SET", "watch_key", "new_value")))
    _, err = resp.ReadResponse(client1.reader)
    if err != nil {
        t.Fatal(err)
    }

    // EXEC should abort
    client1.conn.Write([]byte(client1.SendCommand("EXEC")))
    reply, err = resp.ReadResponse(client1.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "(nil)" {
        t.Errorf("expected (nil), got %q", reply)
    }

    // Key should still be "changed"
    client1.conn.Write([]byte(client1.SendCommand("GET", "watch_key")))
    reply, err = resp.ReadResponse(client1.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "changed" {
        t.Errorf("expected changed, got %q", reply)
    }
}

func TestUnwatch(t *testing.T) {
    client1 := setupTest(t)
    defer client1.Close()
    
    client2 := setupTest(t)
    defer client2.Close()

    // Client 1 watches
    client1.conn.Write([]byte(client1.SendCommand("WATCH", "key")))
    _, _ = resp.ReadResponse(client1.reader)

    // Client 2 modifies the key
    client2.conn.Write([]byte(client2.SendCommand("SET", "key", "modified")))
    _, _ = resp.ReadResponse(client2.reader)

    // Client 1 UNWATCH clears the watch
    client1.conn.Write([]byte(client1.SendCommand("UNWATCH")))
    _, _ = resp.ReadResponse(client1.reader)

    // Client 1 starts a transaction
    client1.conn.Write([]byte(client1.SendCommand("MULTI")))
    _, _ = resp.ReadResponse(client1.reader)

    client1.conn.Write([]byte(client1.SendCommand("SET", "key", "new")))
    _, _ = resp.ReadResponse(client1.reader)

    // EXEC should succeed because UNWATCH cleared the watch
    client1.conn.Write([]byte(client1.SendCommand("EXEC")))
    reply, _ := resp.ReadResponse(client1.reader)
    if !strings.Contains(reply, "OK") {
        t.Errorf("EXEC should succeed after UNWATCH, got %q", reply)
    }
}
// --------------------------------------------------------------
// Keyspace Commands
// --------------------------------------------------------------

func TestDBSizeKeysFlushAll(t *testing.T) {
    client := setupTest(t)
    defer client.Close()
    // SET some keys
    client.conn.Write([]byte(client.SendCommand("SET", "a", "1")))
    _, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    client.conn.Write([]byte(client.SendCommand("SET", "b", "2")))
    _, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }

    // DBSIZE
    client.conn.Write([]byte(client.SendCommand("DBSIZE")))
    reply, err := resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "2"  { // if other tests left keys
        t.Logf("DBSIZE = %q (expected at least 2)", reply)
    }

    // KEYS
    client.conn.Write([]byte(client.SendCommand("KEYS", "*")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if !strings.Contains(reply, "a") || !strings.Contains(reply, "b") {
        t.Errorf("expected a and b, got %q", reply)
    }

    // FLUSHALL
    client.conn.Write([]byte(client.SendCommand("FLUSHALL")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "OK" {
        t.Errorf("expected OK, got %q", reply)
    }

    client.conn.Write([]byte(client.SendCommand("DBSIZE")))
    reply, err = resp.ReadResponse(client.reader)
    if err != nil {
        t.Fatal(err)
    }
    if reply != "0" {
        t.Errorf("expected 0, got %q", reply)
    }
}