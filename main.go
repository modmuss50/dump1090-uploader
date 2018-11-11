package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/textproto"
	"os"
	"time"
)

var (
	server   *string
	port     *string
	dump1090 *string
	mlat     *bool
)

func main() {

	server = flag.String("server", "flights.modmuss50.me", "The remote server hostname or ip")
	port = flag.String("port", "5000", "The remote server port")

	dump1090 = flag.String("dump1090", "localhost", "The dump1090 hostname or ip")
	mlat = flag.Bool("malt", true, "Enables the reading of mlat data from dump1090")

	flag.Parse()

	fmt.Println("Starting dump1090 uploader (ENTER to exit)")

	go connectDump1090()
	if *mlat {
		go connectDump1090mlat()
	}

	go connectRemote()

	waitForExit()
	fmt.Println("Shutting down... have fun :)")

}

//This function connections and maintains the connection to dump1090 and reads the data from it
func connectDump1090() {
	address := *dump1090 + ":30003"
	fmt.Println("Attempting to connect dump1090 @" + address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		dump1090Error(err)
	}
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	defer conn.Close()
	fmt.Println("Connected to dump1090 @" + address)
	for {
		message, err := tp.ReadLine()
		if err != nil {
			dump1090Error(err)
		}
		writeRemote([]byte(message + "\n")) //30 mins to figure out I needed to add back the new line here :D
	}
}

func dump1090Error(err error) {
	fmt.Println("An error occurred when connecting to dump1090, will retry in 10 seconds!")
	fmt.Println(err)
	time.Sleep(10 * time.Second)
	connectDump1090()
}

//Dump1090 uses a different port for mlat aircraft, I am not sure if both 30003 and 30005 are needed
func connectDump1090mlat() {
	address := *dump1090 + ":30005"
	fmt.Println("Attempting to connect dump1090(mlat) @" + address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		dump1090mlatError(err)
	}
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	defer conn.Close()
	fmt.Println("Connected to dump1090(mlat) @" + address)
	for {
		message, err := tp.ReadLine()
		if err != nil {
			dump1090mlatError(err)
		}
		writeRemote([]byte(message + "\n")) //30 mins to figure out I needed to add back the new line here :D
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
	}
}

func remoteError(err error) {
	RemoteServer = nil
	fmt.Println("An error occurred when connecting to remote, will retry in 10 seconds!")
	fmt.Println(err)
	time.Sleep(10 * time.Second)
	connectRemote()
}

//Waits for the enter key to be pressed
func waitForExit() {
	buf := bufio.NewReader(os.Stdin)
	buf.ReadBytes('\n')
}
