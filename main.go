package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

func main() {
	i := 0
	for {
		conn, err := net.Dial("tcp", "127.0.0.1:9000")

		if err != nil {
			log.Fatal(err)
		}

		source := rand.NewSource(time.Now().Unix())
		numero := rand.New(source)
		fmt.Fprintf(conn, strconv.Itoa((numero.Int() % 1e8)))
		fmt.Fprintf(conn, strconv.Itoa((numero.Int() % 1e8)))

		i := i + 1

		fmt.Printf("%d", i)

		time.Sleep(15 * time.Second)
	}
}
