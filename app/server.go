package main

import (
	"bytes"
	"fmt"
	"strings"

	"net"
	"os"
)

type Request struct {
	method  string
	target  string
	version string
}

func (s Request) String() string {
	return fmt.Sprintf("%v %v %v", s.method, s.target, s.version)
}

type Status struct {
	version string
	code    int
	reason  string
}

func (s Status) String() string {
	return fmt.Sprintf("%v %v %v", s.version, s.code, s.reason)
}

type Headers map[string]string

func (h Headers) String() string {
	b := new(bytes.Buffer)
	for k, v := range h {
		fmt.Fprintf(b, "%v %v\r\n", k, v)
	}
	return b.String()
}

type Body string

type HTTPResponse struct {
	status  Status
	headers Headers
	body    Body
}

func (res HTTPResponse) String() string {
	return fmt.Sprintf("%v\r\n%v\r\n%v", res.status, res.headers, res.body)
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	connection, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	response := HTTPResponse{status: Status{version: "HTTP/1.1"}}
	req := make([]byte, 1024)
	connection.Read(req)

	if strings.HasPrefix(string(req), "GET / ") {
		response.status.code = 200
		response.status.reason = "OK"
	} else {
		response.status.code = 404
		response.status.reason = "Not Found"
	}

	connection.Write([]byte(response.String()))

}
