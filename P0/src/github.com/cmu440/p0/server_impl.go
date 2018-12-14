// Implementation of a KeyValueServer. Students should write their code in this file.

package p0

import (
        "fmt"
        "net"
        "os"
        "strconv"
)


type keyValueServer struct {
        port int
}

type ReaderCloser interface {
	Read(p []byte) (n int, err error)
	Close() error
}

// New creates and returns (but does not start) a new KeyValueServer.
func New() KeyValueServer {
	kvs := new(keyValueServer)
	return kvs
}

func (kvs *keyValueServer) Start(port int) error {
	kvs.port = port
        go kvs.StartListeningForConn()

	return nil
}

func (kvs *keyValueServer) StartListeningForConn() {
	ln, err := net.Listen("tcp", ":" + strconv.Itoa(kvs.port))
        if err != nil {
		fmt.Println("Error on listen: ", err)
		os.Exit(-1)
	}

	for {
		fmt.Println("Waiting for inbound connection")
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Couldn't accept: ", err)
			os.Exit(-1)
		}
		go kvs.handleConn(conn)
	}
	fmt.Println("All done")
}

func (kvs *keyValueServer) handleConn(conn ReaderCloser) {
	fmt.Println("Reading once from connection")

	var buf[1024] byte
	n, _ := conn.Read(buf[:])
	fmt.Println("Client sent: ", string(buf[0:n]))
	conn.Close()
}

func (kvs *keyValueServer) Close() {
	// TODO: implement this!
}

func (kvs *keyValueServer) CountActive() int {
	// TODO: implement this!
	return -1
}

func (kvs *keyValueServer) CountDropped() int {
	// TODO: implement this!
	return -1
}

// TODO: add additional methods/functions below!
