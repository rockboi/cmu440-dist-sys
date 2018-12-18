// Implementation of a KeyValueServer. Students should write their code in this file.

package p0

import (
        "fmt"
        "net"
        "os"
        "strconv"
	"strings"
)


type keyValueServer struct {
        port int
	globalreqchan chan *ClientHandle
	writelock chan int
}

type ReaderCloser interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) error
	Close() error
}

type ClientRequestType int

const (
	GETRequest ClientRequestType = 0
	PUTRequest ClientRequestType = 1
	DELRequest ClientRequestType = 2
)

type ClientRequest struct {
	reqType ClientRequestType
	key string
	value string
}

type ClientHandle struct {
	reqChan chan *ClientRequest
	resChan chan string
}

func (clientHandle *ClientHandle) Init() error {
	clientHandle.reqChan = make(chan *ClientRequest)
	clientHandle.resChan = make(chan string)
	return nil
}

// New creates and returns (but does not start) a new KeyValueServer.
func New() KeyValueServer {
	kvs := new(keyValueServer)
	return kvs
}

func (kvs *keyValueServer) Start(port int) error {
	kvs.InitServer(port)
        go kvs.StartListeningForConn()
	go kvs.ProcessRequests()
	return nil
}

func (kvs *keyValueServer) InitServer(port int) error {
	kvs.port = port
	kvs.globalreqchan = make(chan *ClientHandle)
	kvs.writelock = make(chan int, 1)
	kvs.writelock <- 1
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
	clientHandle := new(ClientHandle)
	clientHandle.Init()
	kvs.globalreqchan <- clientHandle
	// n, _ := conn.Read(buf[:])
	// fmt.Println("Client sent: ", string(buf[0:n]))

	for {
		n, err := conn.Read(buf[0:n])
		if err != nil { //EOF, or worse
			break
		}
		clientHandle.reqChan <- ParseClientMessage(string(buf[0:n]))
		conn.Write(<-clientHandle.resChan)
	}

}

func ParseClientMessage(msg string) (clientReq *ClientRequest, err error) {
	// Parse client string here
	vals := strings.split(msg, ",")
	clientReq = new(ClientRequest)
	if (vals[0] == "get") {
		clientReq.reqType = GETRequest
		clientReq.key = vals[1]
	}
	else if (vals[0] == "delete") {
		clientReq.reqType = DELRequest
		clientReq.key = vals[1]
	}
	else if (vals[0] == "put") {
		clientReq.reqType = PUTRequest
		clientReq.key = vals[1]
		clientReq.value = vals[2]
	}
	else {
		// throw error
	}
	return clientReq, err

}

func (kvs *keyValueServer) ProcessRequests() error {
	for clientHandle := range kvs.globalreqchan {
		go kvs.ClientHandler(clientHandle)
	}

}

func (kvs *keyValueServer) ClientHandler(clientHandle *ClientHandle) error {
	for req := range clientHandle.reqChan {
		go kvs.RequestHandler(req, clientHandle.resChan)
	}
}

func ConvertToSingleString(dbvalues []([]byte)) string {
	var finalstr string
	finalstr := ""
	for _, val := range dbvalues {
		finalstr += (string(val[:]) + "\n")
	}
	return finalstr
}

func (kvs *keyValueServer) RequestHandler(clientRequest *ClientRequest, resChan chan string) {
	if (clientRequest.reqType == GETRequest) {
		resChan <- ConvertToSingleString(get(clientRequest.key))
	}
	else if (clientRequest.reqType == PUTRequest) {
		// take lock on server
		<-kvs.writelock
		put(clientRequest.key, clientRequest.value)
		// give back lock
		kvs.writelock <- 1
		resChan <- ""
	}

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
