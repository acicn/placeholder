package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var (
	response        = []byte("PLACEHOLDER")
	responseType    = "text/plain"
	responseLength  = strconv.Itoa(len(response))
	responseHandler = http.HandlerFunc(func(rw http.ResponseWriter, request *http.Request) {
		rw.Header().Set("Content-Type", responseType)
		rw.Header().Set("Content-Length", responseLength)
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(response)
	})
)

func exit(err *error) {
	if *err != nil {
		log.Println("exited with error:", (*err).Error())
		os.Exit(1)
	} else {
		log.Println("exited")
	}
}

func main() {
	var err error
	defer exit(&err)

	var addrs []string

	ports := strings.Split(strings.TrimSpace(os.Getenv("PLACEHOLDER_PORT")), ",")
	for _, port := range ports {
		port = strings.TrimSpace(port)
		if port != "" {
			addrs = append(addrs, ":"+port)
		}
	}
	if len(addrs) == 0 {
		addrs = append(addrs, ":80")
	}

	var servers []*http.Server

	for _, addr := range addrs {
		servers = append(servers, &http.Server{
			Addr:    addr,
			Handler: responseHandler,
		})
	}

	chErr := make(chan error, 1)
	chSig := make(chan os.Signal, 1)
	signal.Notify(chSig, syscall.SIGTERM, syscall.SIGINT)

	for _, server := range servers {
		server := server
		go func() {
			log.Println("listening at", server.Addr)
			chErr <- server.ListenAndServe()
		}()
	}

	defer func() {
		for _, server := range servers {
			log.Println("stop listening at", server.Addr)
			_ = server.Shutdown(context.Background())
		}
	}()

	select {
	case err = <-chErr:
	case sig := <-chSig:
		log.Println("signal caught:", sig.String())
	}
}
