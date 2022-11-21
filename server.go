package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	channels = make(map[string]client)
	private  = make(chan string)
)

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		case msg := <-private:
			msgPvt := strings.SplitN(msg, " ", 4)
			sender := msgPvt[0]
			receiver := msgPvt[2]
			message := msgPvt[3]

			for key, _ := range clients {
				if key == channels[receiver] && receiver != "bot" {

					fmt.Println(sender + " Whispered to " + receiver)
					channels[receiver] <- sender + " Whispered: " + message
					break
				}
				if key == channels[receiver] && receiver == "bot" {

					message = reverse(message)
					channels[sender] <- receiver + " returned: " + message
					break
				}
			}

		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Printf("Read error - %s\n", err)
		}
	}
	username := string(buf[:n])
	fmt.Println(username + " Connected")
	ch <- username + ", welcome to THE chat room"
	messages <- username + " Appeared!"
	entering <- ch
	channels[username] = ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		cmd := strings.Split(input.Text(), " ")
		comando := cmd[0]

		switch comando {
		case "/name":
			messages <- "you are " + cmd[1]
			delete(channels, username)
			username = cmd[1]
			channels[username] = ch

		case "/whisper":
			private <- username + " " + input.Text()
		case "/ls":
			lista := "Online users:\n "
			for key, _ := range channels {
				lista += key + "\n"
			}
			ch <- lista
		case "/quit":
			leaving <- ch
			messages <- username + "Is GONE"
			delete(channels, username)
			return

		case "/help":
			ch <- "\n/name [Newname] : Name Change\n" +
				"/whisper [name] : PM\n" +
				"/ls : list online users\n" +
				"/quit : quit (duh!)\n"

		default:
			fmt.Println("")
			messages <- "[" + username + "]: " + input.Text()
		}
	}
	conn.Close()
}
func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
func reverse(str string) (result string) {
	for _, value := range str {
		result = string(value) + result
	}
	return
}

func main() {

	listener, err := net.Listen("tcp", "localhost:8080")
	fmt.Println("Booting up...")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server is online on port 8080")
	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go handleConn(conn)
	}
}
