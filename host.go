package main

import (
	"net"
	"net/http"
	"github.com/simazhao/go-redis-group/api"
)

func main() {
	serve1()
}

func serve1() {
	eh := make(chan error, 1)
	go func() {
		s := api.NewWebServer()
		defer s.Exit()
		eh <- http.ListenAndServe(":9610", s.Handler)
	}()

	err := <- eh
	println("server end with error:", err.Error())
}

func serve2() {
	if l, err := net.Listen("tcp", ":9610"); err != nil {
		return
	} else {
		eh := make(chan error, 1)
		go func(l net.Listener) {
			h := http.NewServeMux()
			s := api.NewWebServer()
			defer s.Exit()
			h.Handle("/", s.Handler)
			hs := &http.Server{Handler: h}
			eh <- hs.Serve(l)
		}(l)

		err := <- eh
		println("end with error:" + err.Error())
	}
}
