package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"redis/resp"
	"strings"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	stdinReader := bufio.NewReader(os.Stdin)
	connReader := bufio.NewReader(conn)
	nowFormatted := time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")
	fmt.Printf("1:M %s # Server initialized\n", nowFormatted)
	fmt.Printf("1:M %s * Ready to accept connections\n", nowFormatted)
	fmt.Println("listening on localhost:8080")
	for {
		fmt.Print("redis:8080> ")
		cmd, err := stdinReader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			if err == io.EOF {
				conn.Close()
				return
			}
			continue
		}
		quotedCmd := ""
		quoted := 0
		for i := 0; i < len(cmd); i++ {
			if cmd[i] == '"' {
				quoted++
				if quoted == 2 {
					break
				}
				continue
			}
			if quoted == 1 {
				quotedCmd += string(cmd[i])
			}
		}
		cmdTab := strings.Fields(cmd)
		if quoted == 2 {
			cmdTab = []string{cmdTab[0], cmdTab[1], quotedCmd}
		}
		if len(cmdTab) == 0 {
			continue
		}
		if strings.ToUpper(cmdTab[0]) == "Q" {
			conn.Close()
			return
		}
		if strings.ToUpper(cmdTab[0]) == "CLEAR" {
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				fmt.Printf("Command failed: %v\n", err)
			}
			continue
		}
		conn.Write(resp.EncodeCommand(cmdTab))
		response, readErr := resp.ReadResponse(connReader)
		if response != "" {
			fmt.Println(response)
		}
		if readErr != nil {
			if readErr == io.EOF {
				fmt.Println("server disconnected")
			} else {
				fmt.Println(readErr)
			}
			conn.Close()
			return
		}
	}
}
