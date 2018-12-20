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
	activeconns int
	droppedconns int
}

type ReaderWriterCloser interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

type ClientRequestType int

const (
	GETRequest ClientRequestType = 0
	PUTRequest ClientRequestType = 1
	DELRequest ClientRequestType = 2
	SlowClientsBuffer int = 500
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
	clientHandle.resChan = make(chan string, SlowClientsBuffer)
	return nil
}

// New creates and returns (but does not start) a new KeyValueServer.
func New() KeyValueServer {
	kvs := new(keyValueServer)
	return kvs
}

func (kvs *keyValueServer) Start(port int) error {
	init_db()
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
	kvs.activeconns = 0
	kvs.droppedconns = 0
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

func (kvs *keyValueServer) handleConn(conn ReaderWriterCloser) {
	kvs.activeconns++
	fmt.Println("Reading once from connection")

	var buf[1024] byte
	clientHandle := new(ClientHandle)
	clientHandle.Init()
	kvs.globalreqchan <- clientHandle
	// n, _ := conn.Read(buf[:])
	// fmt.Println("Client sent: ", string(buf[0:n]))

	for {
		n, err := conn.Read(buf[:])
		if err != nil { //EOF, or worse
			break
		}
		clientReq, err := ParseClientMessage(string(buf[0:n]))
		if err != nil {
			fmt.Println("Could not parse message: ", err)
			conn.Write([]byte("Could not parse message\n"))
			continue
		}
		clientHandle.reqChan <- clientReq
		clientRes := <-clientHandle.resChan
		conn.Write([]byte(clientRes))
	}
	kvs.activeconns--
	kvs.droppedconns++
}

func ParseClientMessage(msg string) (clientReq *ClientRequest, err error) {
	// Parse client string here
	var cleanmsg = strings.TrimSuffix(msg, "\n")
	vals := strings.Split(cleanmsg, ",")
	clientReq = new(ClientRequest)
	if (strings.Compare(vals[0], "get") == 0) {
		clientReq.reqType = GETRequest
		clientReq.key = vals[1]
	} else if (strings.Compare(vals[0], "delete") == 0) {
		clientReq.reqType = DELRequest
		clientReq.key = vals[1]
	} else if (strings.Compare(vals[0], "put") == 0) {
		clientReq.reqType = PUTRequest
		clientReq.key = vals[1]
		clientReq.value = vals[2]
	} else {
		return nil, fmt.Errorf("Unknown request type %s", vals[0])
	}
	return clientReq, nil

}

func (kvs *keyValueServer) ProcessRequests() error {
	for clientHandle := range kvs.globalreqchan {
		go kvs.ClientHandler(clientHandle)
	}
	return nil
}

func (kvs *keyValueServer) ClientHandler(clientHandle *ClientHandle) error {
	for req := range clientHandle.reqChan {
		go kvs.RequestHandler(req, clientHandle.resChan)
	}
	return nil
}

func ConvertToSingleString(dbvalues []([]byte)) string {
	finalstr := ""
	for _, val := range dbvalues {
		finalstr += (string(val[:]) + "\n")
	}
	return finalstr
}

func (kvs *keyValueServer) RequestHandler(clientRequest *ClientRequest, resChan chan string) error {
	if (clientRequest.reqType == GETRequest) {
		// get all values associated with key
		respStr := ConvertToSingleString(get(clientRequest.key))
		fmt.Printf("GET request, key: %s, value: %s.\n", clientRequest.key, respStr)
		resChan <- respStr
	} else if (clientRequest.reqType == PUTRequest) {
		// take lock on server
		<-kvs.writelock
		fmt.Printf("PUT request, key: %s, value: %s.\n", clientRequest.key, clientRequest.value)
		put(clientRequest.key, []byte(clientRequest.value))
		// give back lock
		kvs.writelock <- 1
		resChan <- ""
	} else if (clientRequest.reqType == DELRequest) {
		//delete key
		fmt.Printf("DELETE request, key: %s.\n", clientRequest.key)
		clear(clientRequest.key)
		resChan <- ""
	} else {
		return fmt.Errorf("Unsupported request type %g", clientRequest.reqType)
	}
	return nil
}

func (kvs *keyValueServer) Close() {
	// TODO
	// Signal all channels to close?
}

func (kvs *keyValueServer) CountActive() int {
	return kvs.activeconns
}

func (kvs *keyValueServer) CountDropped() int {
	return kvs.droppedconns
}
