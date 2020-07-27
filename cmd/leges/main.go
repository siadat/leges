package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/siadat/leges/httpserver"
)

func main() {
	var (
		optsAddr       = flag.String("addr", ":5120", "HTTP bind address")
		optsPolicyFile = flag.String("policies", "policies.yaml", "Policy file")
	)
	flag.Parse()

	yamlFile, err := os.Open(*optsPolicyFile)
	if err != nil {
		panic(err)
	}

	policies, err := httpserver.LoadPoliciesFromYaml(yamlFile)
	if err != nil {
		panic(err)
	}

	srv := http.Server{
		Addr:    *optsAddr,
		Handler: &httpserver.Server{Policies: policies},
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Printf("Caught ctrl-c...")
		if err := srv.Shutdown(context.TODO()); err != nil {
			panic(err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("Server starting on %s", *optsAddr)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		panic(err)
	}
	<-idleConnsClosed
	log.Printf("à bientôt!")
}
