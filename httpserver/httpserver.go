package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/siadat/leges"
	"gopkg.in/yaml.v2"
)

type Response map[string]interface{}

type Server struct {
	Policies []leges.Policy
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.String())

	objectAttributes, err := UnmarshalAttributes(r.URL.Query().Get("object"))
	if err != nil {
		fmt.Fprint(w, MustMarshal(Response{
			"error": fmt.Sprintf("JSON parse error: 'object' must be valid JSON: %s", err.Error()),
		}))
		return
	}

	subjectAttributes, err := UnmarshalAttributes(r.URL.Query().Get("subject"))
	if err != nil {
		fmt.Fprint(w, MustMarshal(Response{
			"error": fmt.Sprintf("JSON parse error: 'subject' must be valid JSON: %s", err.Error()),
		}))
		return
	}

	request := leges.Request{
		Object:  objectAttributes,
		Subject: subjectAttributes,
		Action:  r.URL.Query().Get("action"),
	}

	lg, err := leges.NewLeges(srv.Policies)
	if err != nil {
		panic(err)
	}
	ok, policy, err := lg.Match(request, nil)

	if err != nil {
		fmt.Fprint(w, MustMarshal(Response{
			"error": err.Error(),
		}))
		return
	}

	if policy != nil {
		fmt.Fprint(w, MustMarshal(Response{
			"id":    policy.ID,
			"match": ok,
		}))
		return
	}

	{
		fmt.Fprint(w, MustMarshal(Response{
			"match": ok,
		}))
	}
}

func LoadPoliciesFromYaml(y io.Reader) ([]leges.Policy, error) {
	decoder := yaml.NewDecoder(y)
	var policies []leges.Policy
	err := decoder.Decode(&policies)
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func UnmarshalAttributes(jsonified string) (leges.Attributes, error) {
	var attributes leges.Attributes
	buf := bytes.NewBufferString(jsonified)
	err := json.NewDecoder(buf).Decode(&attributes)
	if err != nil {
		return nil, err
	}
	return attributes, nil
}

func MustMarshal(whatever interface{}) string {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(whatever)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
