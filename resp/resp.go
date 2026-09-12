package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	maxBulkPayload   = 512 * 1024 * 1024
	maxArrayElements = 1_000_000
)

func EncodeCommand(cmds []string) []byte {
	var response []byte
	var strResponse strings.Builder
	fmt.Fprintf(&strResponse, "*%d\r\n", len(cmds))
	for i := range cmds {
		fmt.Fprintf(&strResponse, "$%d\r\n", len(cmds[i]))
		fmt.Fprintf(&strResponse, "%s\r\n", cmds[i])
	}
	response = []byte(strResponse.String())
	return response
}
func ReadCommand(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	cmdSize, err := GetCmdSize(line)
	if err != nil {
		return nil, err
	}
	nbrCmd, err := strconv.Atoi(cmdSize)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if nbrCmd < 0 || nbrCmd > maxArrayElements {
		return nil, fmt.Errorf("invalid command count: %d", nbrCmd)
	}
	var cmds []string
	for i := 0; i < nbrCmd; i++ {
		size, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		sizeHeader, err := GetCmdSize(size)
		if err != nil {
			return nil, err
		}
		sizeVal, err := strconv.Atoi(sizeHeader)
		if err != nil {
			fmt.Println("failed conversion : ", err)
			return nil, err
		}
		if sizeVal < 0 || sizeVal > maxBulkPayload {
			return nil, fmt.Errorf("invalid bulk string size: %d", sizeVal)
		}
		cmdBuf := make([]byte, sizeVal)
		_, err = io.ReadFull(reader, cmdBuf)
		if err != nil {
			return nil, err
		}
		cmdVal := string(cmdBuf)
		cmds = append(cmds, cmdVal)
		trailer, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if trailer != "\r\n" {
   		 return nil, fmt.Errorf("invalid bulk string terminator: %q", trailer)
		}
	}
	return cmds, nil
}
func GetCmdSize(chunk string) (string, error) {
	strTab := strings.Split(chunk, "\r\n")
	valStr := strTab[0]
	if len(valStr) < 2 {
		return "", fmt.Errorf("invalid RESP header: %q", chunk)
	}
	if valStr[0] != '*' && valStr[0] != '$' {
		return "", fmt.Errorf("invalid RESP header prefix: %q", valStr[0])
	}
	return valStr[1:], nil
}

func ErrorMsg(size int, num int, cmd string, op string) []byte {
	var valid bool
	switch op {
	case "==":
		valid = size == num
	case ">=":
		valid = size >= num
	default:
		return fmt.Appendf(nil, "-ERR internal error: invalid operator '%s' for command '%s'\r\n", op, cmd)
	}
	if !valid {
		strResponse := fmt.Sprintf("-ERR wrong number of arguments for '%s' command\r\n", cmd)
		return []byte(strResponse)
	}
	return nil
}

func Ping() []byte {
	return []byte("+PONG\r\n")
}
func QueuedResponse() []byte {
	return []byte("+QUEUED\r\n")
}
func OkResponse() []byte {
	return []byte("+OK\r\n")
}
func IntResponse(n int) []byte {
	return fmt.Appendf(nil, ":%d\r\n", n)
}

func BulkResponse(s string) []byte {
	return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(s), s)
}

func NullResponse() []byte {
	return []byte("$-1\r\n")
}

func EmptyArrayResponse() []byte {
	return []byte("*0\r\n")
}
func ArrayResponse(n int) []byte {
	return fmt.Appendf(nil, "*%d\r\n", n)
}

func ErrorResponse(msg string) []byte {
	return fmt.Appendf(nil, "-ERR %s\r\n", msg)
}

func ReadResponse(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(line, "\r\n") {
		return "", fmt.Errorf("invalid RESP line terminator")
	}
	line = strings.TrimSuffix(line, "\r\n")
	if line == "" {
		return "", fmt.Errorf("empty response")
	}

	switch line[0] {
	case '+':
		// Simple string: +OK\r\n → "OK"
		return line[1:], nil

	case '-':
		// Error: -ERR something\r\n → "ERR: something"
		return "ERR: " + line[1:], nil

	case ':':
		// Integer: :123\r\n → "123"
		return line[1:], nil

	case '$':
		// Bulk string or null
		if line == "$-1" {
			return "(nil)", nil
		}
		var length int
		parsed, err := fmt.Sscanf(line, "$%d", &length)
		if err != nil || parsed != 1 || length < -1 {
			return "", fmt.Errorf("invalid bulk string length: %s", line)
		}
		if length == -1 {
			return "(nil)", nil
		}
		data := make([]byte, length)
		_, err = io.ReadFull(reader, data)
		if err != nil {
			return "", err
		}
		// Consume trailing \r\n
		reader.ReadString('\n')
		return string(data), nil

	case '*':
		// Array: *N\r\n...
		var count int
		parsed, err := fmt.Sscanf(line, "*%d", &count)
		if err != nil || parsed != 1 || count < -1 {
			return "", fmt.Errorf("invalid array count: %s", line)
		}
		if count == -1 {
			return "(nil)", nil
		}
		if count == 0 {
			return "[]", nil
		}
		// Collect all elements
		elements := make([]string, count)
		for i := 0; i < count; i++ {
			// Read each element recursively
			elem, err := ReadResponse(reader)
			if err != nil {
				return "", err
			}
			elements[i] = elem
		}
		return fmt.Sprintf("[%s]", strings.Join(elements, " ")), nil

	default:
		return "", fmt.Errorf("unknown response: %s", line)
	}
}
