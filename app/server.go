package main

import (
	"bytes"
	"compress/gzip"
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

func notFoundController() (response HTTPResponse) {
	response.status = Status{version: "HTTP/1.1", code: 404, reason: "Not Found"}
	return
}

func rootController() (response HTTPResponse) {
	response.status = Status{version: "HTTP/1.1", code: 200, reason: "OK"}
	return
}

func echoController(req HTTPRequest) (response HTTPResponse) {
	response.status.version = "HTTP/1.1"

	p := strings.Split(req.request.target, "/")
	if len(p) != 3 {
		response.status.code = 404
		response.status.reason = "Not Found"
		return
	}

	response.body = Body(p[2])

	response.headers = Headers{"Content-Type": "text/plain"}

	acceptedEncodings := strings.Split(req.headers["Accept-Encoding"], ",")

	for _, encoding := range acceptedEncodings {
		if strings.TrimSpace(encoding) == "gzip" {
			response.headers["Content-Encoding"] = "gzip"

			var bodyBytes bytes.Buffer
			writer := gzip.NewWriter(&bodyBytes)
			writer.Write([]byte(response.body))
			writer.Close()
			response.body = Body(bodyBytes.String())
			break
		}
	}

	response.status.code = 200
	response.status.reason = "OK"
	response.headers["Content-Length"] = fmt.Sprint(len(response.body))

	return
}

func userAgentController(req HTTPRequest) (response HTTPResponse) {
	response.status.version = "HTTP/1.1"

	userAgent, exists := req.headers["User-Agent"]
	if !exists {
		response.status.code = 400
		response.status.reason = "Invalid Headers"
		return
	}

	response.status.code = 200
	response.status.reason = "OK"
	response.body = Body(userAgent)
	headers := Headers{"Content-Type": "text/plain", "Content-Length": fmt.Sprint(len(response.body))}
	response.headers = headers
	return
}

func filesController(req HTTPRequest) (response HTTPResponse) {
	response.status.version = "HTTP/1.1"
	response.status.code = 404
	response.status.reason = "Not Found"

	p := strings.Split(req.request.target, "/")
	if len(p) != 3 {
		fmt.Println(fmt.Printf("Invalid route: %v", req.request.target))
		return
	}
	fileName := p[2]

	dir := "files/"

	switch req.request.method {
	case "GET":
		file, err := os.ReadFile(dir + fileName)
		if err != nil {
			fmt.Println("Failed to read file: ", err.Error())
			return
		}

		headers := Headers{"Content-Type": "application/octet-stream", "Content-Length": fmt.Sprint(len(file))}

		response.status.code = 200
		response.status.reason = "OK"
		response.headers = headers
		response.body = Body(file)

	case "POST":
		file, err := os.Create(dir + fileName)
		if err != nil {
			fmt.Println("Failed to create file: ", err.Error())
			response.status.code = 500
			response.status.reason = "Internal Server Error"
			return
		}

		defer file.Close()

		content := strings.TrimRight(string(req.body), "\x00")

		_, werr := file.WriteString(content)
		if werr != nil {
			fmt.Println("Failed to write file: ", werr.Error())
			response.status.code = 500
			response.status.reason = "Internal Server Error"
			return
		}

		response.status.code = 201
		response.status.reason = "Created"

	default:
		response.status.code = 405
		response.status.reason = "Method Not Allowed"
	}

	return
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(connection)
	}
}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	req := make([]byte, 1024)
	_, err := connection.Read(req)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
	}

	request, err := ParseHTTPRequest(req)
	if err != nil {
		fmt.Println("Error parsing HTTP request: ", err.Error())
	}

	var response HTTPResponse

	if request.request.target == "/" {
		response = rootController()
	} else if strings.HasPrefix(request.request.target, "/echo") {
		response = echoController(request)
	} else if strings.HasPrefix(request.request.target, "/user-agent") {
		response = userAgentController(request)
	} else if strings.HasPrefix(request.request.target, "/files") {
		response = filesController(request)
	} else {
		response = notFoundController()
	}

	connection.Write([]byte(response.String()))
}
