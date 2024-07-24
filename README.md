# Simple Go HTTP Server

This is my implementation of [Codecrafters - Build your own HTTP server](https://app.codecrafters.io/courses/http-server/overview) challenge.

## Motivation

I have decided to take this challenge since I am in process of learning Go and also wanted to deepen my understanding of how HTTP servers work.

###### *aaand the course was free at the time...*

## Features

The implementation contains custom structs for HTTP request and response and functions to parse and serialize them.

The server can handle multiple requests concurrently and can accept and serve files from the '/files' directory.

Endpoints:

- **GET '/'**: returns a simple response with a status code of 200
- **GET '/user-agent'**: reads the User-Agent header and returns it in the response body
- **GET '/echo/{*message*}'**: returns a response with the *message* in body and appropriate headers.
  >*message* is 'gzip' compressed if request header 'Accept-Encoding' contains 'gzip'.

- **GET '/files/{*filename*}'**: returns a response with the file content in the body and appropriate headers
    >The server will look for the file in the '/files' directory. If the file is not found, a 404 status code will be returned.
- **POST '/files/{*filename*}'**: saves the request body to a file with the name *filename* in the '/files' directory and returns a 201 status code if successful.
    >If the file already exists, it will be overwritten.


Unsupported routes will return a 404 status code.

## Usage

To run the server, execute the following command:

```bash
go run app/server.go
```

The server will start on port 4221.
