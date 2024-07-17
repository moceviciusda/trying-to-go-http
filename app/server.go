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
		fmt.Fprintf(b, "%v: %v\r\n", k, v)
	}
	return b.String()
}

type Body string

type HTTPResponse struct {
	status  Status
	headers Headers
	body    Body
}

type HTTPRequest struct {
	request Request
	headers Headers
	body    Body
}

func (res HTTPResponse) String() string {
	return fmt.Sprintf("%v\r\n%v\r\n%v", res.status, res.headers, res.body)
}

func ParseHTTPRequest(req []byte) (HTTPRequest, error) {
	var result HTTPRequest
	sreq, srest, ok := strings.Cut(string(req), "\r\n")

	sreqSlice := strings.Split(sreq, " ")

	if !ok || len(sreqSlice) < 3 {
		return result, fmt.Errorf("invalid request: %v", req)
	}
	result.request = Request{method: sreqSlice[0], target: sreqSlice[1], version: sreqSlice[2]}

	sheaders, sbody, ok := strings.Cut(srest, "\r\n\r\n")
	if !ok {
		return result, fmt.Errorf("invalid request: %v", req)
	}

	headers := map[string]string{}
	for _, sheader := range strings.Split(sheaders, "\r\n") {
		k, v, ok := strings.Cut(sheader, ": ")
		if ok {
			headers[k] = v
		}
	}

	result.headers = headers

	result.body = Body(sbody)

	return result, nil
}

func echoController(req HTTPRequest) (response HTTPResponse) {
	response.status.version = "HTTP/1.1"

	p := strings.Split(req.request.target, "/")
	if len(p) != 3 {
		response.status.code = 404
		response.status.reason = "Not Found"
	} else {
		body := Body(p[2])
		headers := Headers{"Content-Type": "text/plain", "Content-Length": fmt.Sprint(len(body))}
		print(headers.String())

		response.status.code = 200
		response.status.reason = "OK"
		response.headers = headers
		response.body = body
	}

	return
}

func notFoundController() (response HTTPResponse) {
	response.status = Status{version: "HTTP/1.1", code: 404, reason: "Not Found"}
	return
}

func rootController() (response HTTPResponse) {
	response.status = Status{version: "HTTP/1.1", code: 200, reason: "OK"}
	return
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

	req := make([]byte, 1024)
	connection.Read(req)

	request, err := ParseHTTPRequest(req)
	if err != nil {
		fmt.Println("Error parsing HTTP request: ", err.Error())
		os.Exit(1)
	}

	var response HTTPResponse
	// TODO: figure out how to do proper routing. This is 🤮
	if request.request.target == "/" {
		response = rootController()
	} else if strings.HasPrefix(request.request.target, "/echo") {
		response = echoController(request)
	} else {
		response = notFoundController()
	}

	print(response.String())
	connection.Write([]byte(response.String()))

}
