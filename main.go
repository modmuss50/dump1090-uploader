package main

import (
	"bufio"
	"code.cloudfoundry.org/bytefmt"
	"flag"
	"fmt"
	"net"
	"net/textproto"
	"os"
	"time"
)

var (
	server       *string
	port         *string
	dump1090     *string
	dump1090Port *string
	mlat         *bool
	mlatPort     *string
)

var (
	messageCount             = 0
	dump1090Count            = 0
	dump1090mlatCount        = 0
	messageSize       uint64 = 0
)

func main() {

	server = flag.String("server", "flights.modmuss50.me", "The remote server hostname or ip")
	port = flag.String("port", "5000", "The remote server port")

	dump1090 = flag.String("dump1090", "localhost", "The dump1090 hostname or ip")
	dump1090Port = flag.String("dump1090Port", "30005", "The dump1090 raw output port")
	mlat = flag.Bool("mlat", true, "Enables the reading of mlat data from dump1090")
	mlatPort = flag.String("mlatPort", "30105", "The dump1090 raw output port for mlat data")
	flag.Parse()

	fmt.Println("Starting dump1090 uploader (ENTER to exit)")

	go connectDump1090()
	if *mlat {
		go connectDump1090mlat()
	}

	go connectRemote()

	go printDebug()

	waitForExit()
	fmt.Println("Shutting down... have fun :)")

}

//This function connections and maintains the connection to dump1090 and reads the data from it
func connectDump1090() {
	dump1090Count = 0
	address := *dump1090 + ":" + *dump1090Port
	fmt.Println("Attempting to connect dump1090 @" + address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		dump1090Error(err)
		conn.Close()
		return
	}
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	defer conn.Close()
	fmt.Println("Connected to dump1090 @" + address)
	go dump1090KeepAlive(conn)
	for {
		message, err := tp.ReadLine()
		if err != nil {
			dump1090Error(err)
			conn.Close()
			return
		}
		writeRemote([]byte(message + "\n")) //30 mins to figure out I needed to add back the new line here :D
		dump1090Count++
	}
}

func dump1090Error(err error) {
	fmt.Println("An error occurred when connecting to dump1090, will retry in 10 seconds!")
	fmt.Println(err)
	time.Sleep(10 * time.Second)
	connectDump1090()
}

//If there were 0 messages from dump1090 reconnect as some times the connection can close, its a work around that fixes a bug
func dump1090KeepAlive(conn net.Conn) {
	dump1090Count = 0
	time.Sleep(120 * time.Second)
	if dump1090Count == 0 {
		fmt.Println("No messages from dump1090 in the last 60 seconds, reconnecting...")
		conn.Close()
		connectDump1090()
	} else {
		dump1090KeepAlive(conn)
	}
}

//Dump1090 uses a different port for mlat aircraft, I am not sure if both 30003 and 30005 are needed
func connectDump1090mlat() {
	dump1090mlatCount = 0
	address := *dump1090 + ":" + *mlatPort
	fmt.Println("Attempting to connect dump1090(mlat) @" + address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		dump1090mlatError(err)
		conn.Close()
		return
	}
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	defer conn.Close()
	fmt.Println("Connected to dump1090(mlat) @" + address)
	go dump1090MlatKeepAlive(conn)
	for {
		message, err := tp.ReadLine()
		if err != nil {
			dump1090mlatError(err)
			conn.Close()
			return
		}
		writeRemote([]byte(message + "\n")) //30 mins to figure out I needed to add back the new line here :D
		dump1090mlatCount++
	}
}

//If there were 0 messages from dump1090(mlat) reconnect as some times the connection can close, its a work around that fixes a bug
func dump1090MlatKeepAlive(conn net.Conn) {
	dump1090mlatCount = 0
	time.Sleep(120 * time.Second)
	if dump1090mlatCount == 0 {
		fmt.Println("No messages from dump1090(mlat) in the last 60 seconds, reconnecting...")
		conn.Close()
		connectDump1090mlat()
	} else {
		dump1090MlatKeepAlive(conn)
	}
}

func dump1090mlatError(err error) {
	fmt.Println("An error occurred when connecting to dump1090(mlat), will retry in 10 seconds!")
	fmt.Println(err)
	time.Sleep(10 * time.Second)
	connectDump1090mlat()
}

var (
	RemoteServer net.Conn
)

//This function connects and maintains the connection to the remote server, this pushes the data from dump1090 to the remote server
func connectRemote() {
	address := *server + ":" + *port
	fmt.Println("Attempting to connect to remote @" + address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		remoteError(err)
	}
	fmt.Println("Connected to remote @" + address)
	RemoteServer = conn

	//TODO reply to the keep alive messages?

}

func remoteConnected() bool {
	return RemoteServer != nil
}

func writeRemote(message []byte) {
	if !remoteConnected() {
		return
	}
	_, err := RemoteServer.Write(message)
	if err != nil {
		remoteError(err)
	} else {
		messageCount++
		messageSize += uint64(len(message))
	}
}

func remoteError(err error) {
	RemoteServer = nil
	fmt.Println("An error occurred when connecting to remote, will retry in 10 seconds!")
	fmt.Println(err)
	time.Sleep(10 * time.Second)
	connectRemote()
}

func printDebug() {
	time.Sleep(60 * time.Second)
	fmt.Printf("%d messages sent in the last 60 seconds (%sB) \n", messageCount, bytefmt.ByteSize(messageSize))
	messageCount = 0
	messageSize = 0
	printDebug()
}

//Waits for the enter key to be pressed
func waitForExit() {
	buf := bufio.NewReader(os.Stdin)
	buf.ReadBytes('\n')
}
