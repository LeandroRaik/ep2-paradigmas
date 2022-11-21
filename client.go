package main

// Tamo acabando viado, bora pra cima Corinthians!
import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	fmt.Println("Connected!")
	fmt.Println("\nType /help for a list of available commands\n")
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan struct{})
	fmt.Print("Who are you?\n> ")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	conn.Write([]byte(input.Text()))

	go func() {
		io.Copy(os.Stdout, conn)
		log.Println("done")
		done <- struct{}{} // sinaliza para a gorrotina principal
	}()
	mustCopy(conn, os.Stdin)
	conn.Close()
	<-done // espera a gorrotina terminar
}
