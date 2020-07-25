package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/siadat/leges"
	"gopkg.in/yaml.v2"
)

func main() {
	var (
		optsAddr       = flag.String("addr", ":5120", "HTTP bind address")
		optsPolicyFile = flag.String("policies", "policies.yaml", "Policy file")
	)

	type Response map[string]interface{}

	flag.Parse()

	policies, err := loadPoliciesFromYaml(*optsPolicyFile)
	if err != nil {
		panic(err)
	}

	srv := http.Server{
		Addr: *optsAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Printf("%s %s", r.Method, r.URL.String())

			objectAttributes, err := unmarshalAttributes(r.URL.Query().Get("object"))
			if err != nil {
				fmt.Fprintf(w, mustMarshal(Response{
					"err": fmt.Sprintf("JSON parse error: 'subject' must be valid JSON: %s", err.Error()),
				}))
				return
			}

			subjectAttributes, err := unmarshalAttributes(r.URL.Query().Get("subject"))
			if err != nil {
				fmt.Fprintf(w, mustMarshal(Response{
					"err": fmt.Sprintf("JSON parse error: 'subject' must be valid JSON: %s", err.Error()),
				}))
				return
			}

			request := leges.Request{
				Object:  objectAttributes,
				Subject: subjectAttributes,
				Action:  r.URL.Query().Get("action"),
			}

			ok, policy, err := leges.Match(policies, request, nil)

			if err != nil {
				fmt.Fprintf(w, mustMarshal(Response{
					"error": err.Error(),
				}))
				return
			}

			if policy != nil {
				fmt.Fprintf(w, mustMarshal(Response{
					"id":    policy.ID,
					"match": ok,
				}))
				return
			}

			{
				fmt.Fprintf(w, mustMarshal(Response{
					"match": ok,
				}))
			}
		}),
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

func unmarshalAttributes(jsonified string) (leges.Attributes, error) {
	var attributes leges.Attributes
	buf := bytes.NewBufferString(jsonified)
	err := json.NewDecoder(buf).Decode(&attributes)
	if err != nil {
		return nil, err
	}
	return attributes, nil
}

func mustMarshal(whatever interface{}) string {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(whatever)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func loadPoliciesFromYaml(filepath string) ([]leges.Policy, error) {
	yamlFile, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(yamlFile)
	var policies []leges.Policy
	err = decoder.Decode(&policies)
	if err != nil {
		return nil, err
	}
	return policies, nil
}
