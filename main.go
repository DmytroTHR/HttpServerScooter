package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	httpServer := &http.Server{
		Handler: &requestHandler{},
		Addr:    ":8888",
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		log.Println("Server shutting down..")
		close(idleConnsClosed)
	}()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}

type requestHandler struct {
}

func (rh *requestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		defer r.Body.Close()
		queryStr, err := bodyToString(r.Body)
		if err != nil {
			fmt.Fprintf(w, "Unable to convert body to string: %v", err)
			return
		}

		workWithRequest(queryStr, w)
	} else {
		fmt.Fprintln(w, "Server is working...")
	}
}

func bodyToString(closer io.ReadCloser) (string, error)  {
	bodyBytes, err := ioutil.ReadAll(closer)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func workWithRequest(query string, w http.ResponseWriter)  {
	var request struct {
		Method string
		Parameters interface{}
	}
	decoder := json.NewDecoder(strings.NewReader(query))
	err := decoder.Decode(&request)
	if err != nil {
		log.Panicf("error decoding request: %v", err)
	}

	switch request.Method {
	case "CreateUser":
		
	}
	
	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(request)
}