package resp

import (
	"fmt"
	"strings"
	"strconv"
	"bufio"
	"io"
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
	nbrCmd, err := strconv.Atoi(GetCmdSize(line))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var cmds []string
	for i := 0; i < nbrCmd; i++ {
		size, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		sizeVal, err := strconv.Atoi(GetCmdSize(size))
		if err != nil {
			fmt.Println("failed conversion : ", err)
			return nil, err
		}
		cmdBuf := make([]byte, sizeVal)
		_, err = io.ReadFull(reader, cmdBuf)
		if err != nil {
			return nil, err
		}
		cmdVal := string(cmdBuf)
		cmds = append(cmds, cmdVal)
		_, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
	}
	return cmds, nil
}
func GetCmdSize(chunk string) string {
	strTab := strings.Split(chunk, "\r\n")
	valStr := strTab[0]
	return valStr[1:]
}


func ErrorMsg(size int, num int, cmd string, op string)([]byte){
	var valid bool
	switch op {
	case "==":
		valid = size == num
	case ">=":
		valid = size >= num
	default:
		return fmt.Appendf(nil, "-ERR internal error: invalid operator '%s' for command '%s'\r\n", op, cmd)	}
 	if !valid {
		strResponse := fmt.Sprintf("-ERR wrong number of arguments for '%s' command\r\n",cmd)
		return []byte(strResponse)  
	}
	return nil
}

func Ping()[]byte{
	return []byte("+PONG\r\n")
}
func QueuedResponse()[]byte{
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
    return fmt.Appendf(nil, "*%d\r\n",n)
}

func ErrorResponse(msg string) []byte {
    return fmt.Appendf(nil, "-ERR %s\r\n", msg)
}

func ReadResponse(reader *bufio.Reader) (string, error) {
    line, err := reader.ReadString('\n')
    if err != nil {
        return "", err
    }
    line = strings.TrimSpace(line)
    if line == "" {
        return "", fmt.Errorf("empty response")
    }

    switch line[0] {
    case '+':
        // Simple string: +OK\r\n → "OK"
        return strings.TrimSuffix(line[1:], "\r\n"), nil

    case '-':
        // Error: -ERR something\r\n → "ERR: something"
        return "ERR: " + strings.TrimSuffix(line[1:], "\r\n"), nil

    case ':':
        // Integer: :123\r\n → "123"
        return strings.TrimSuffix(line[1:], "\r\n"), nil

    case '$':
        // Bulk string or null
        if line == "$-1" {
            return "(nil)", nil
        }
        var length int
        fmt.Sscanf(line, "$%d", &length)
        if length == -1 {
            return "(nil)", nil
        }
        data := make([]byte, length)
        _, err := io.ReadFull(reader, data)
        if err != nil {
            return "", err
        }
        // Consume trailing \r\n
        reader.ReadString('\n')
        return string(data), nil

    case '*':
        // Array: *N\r\n...
        var count int
        fmt.Sscanf(line, "*%d", &count)
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