package httpserver_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/siadat/leges"
	"github.com/siadat/leges/httpserver"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	srv := httptest.NewServer(&httpserver.Server{Policies: []leges.Policy{
		{
			ID:        "policy0",
			Condition: "object == subject",
			Actions:   []string{"ACTION0"},
		},
		{
			ID:        "policy1",
			Condition: "object.k == subject.k",
			Actions:   []string{"ACTION1"},
		},
	}})
	defer srv.Close()

	testCases := []struct {
		subject        string
		object         string
		action         string
		expectedResult string
	}{
		{
			subject:        httpserver.MustMarshal(leges.Attributes{"m": "1", "n": "2"}),
			object:         httpserver.MustMarshal(leges.Attributes{"m": "1", "n": "2"}),
			action:         "ACTION0",
			expectedResult: `{"match": true, "id": "policy0"}`,
		},
		{
			subject:        httpserver.MustMarshal(leges.Attributes{"k": "v"}),
			object:         httpserver.MustMarshal(leges.Attributes{"k": "v"}),
			action:         "ACTION1",
			expectedResult: `{"match": true, "id": "policy1"}`,
		},
		{
			subject:        "",
			object:         httpserver.MustMarshal(leges.Attributes{"k": "v"}),
			action:         "ACTION1",
			expectedResult: `{"error": "JSON parse error: 'subject' must be valid JSON: EOF"}`,
		},
		{
			subject:        httpserver.MustMarshal(leges.Attributes{"k": "v"}),
			object:         "",
			action:         "ACTION1",
			expectedResult: `{"error": "JSON parse error: 'object' must be valid JSON: EOF"}`,
		},
		{
			subject:        httpserver.MustMarshal(leges.Attributes{"k": "v"}),
			object:         httpserver.MustMarshal(leges.Attributes{"k": "v"}),
			action:         "ACTION2",
			expectedResult: `{"match": false}`,
		},
	}

	for _, tt := range testCases {
		req, err := http.NewRequest("GET", srv.URL, nil)
		require.NoError(t, err)

		params := req.URL.Query()
		params.Set("subject", tt.subject)
		params.Set("object", tt.object)
		params.Set("action", tt.action)
		req.URL.RawQuery = params.Encode()

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		resBody, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.JSONEq(t, tt.expectedResult, string(resBody))
	}
}

func TestLoadPoliciesFromYaml(t *testing.T) {
	policies, err := httpserver.LoadPoliciesFromYaml(bytes.NewBufferString(`
  - id: admin_can_update_and_view_pages
    condition: |
      subject.role == "admin"
      and object.type in ["page", "adminpage"]
    actions:
      - VIEW
      - UPDATE

  - id: guest_can_only_view_pages
    condition: |
      subject.role == "guest"
      and object.type == "page"
    actions:
      - VIEW
`))
	if err != nil {
		panic(err)
	}
	require.NoError(t, err)
	require.Len(t, policies, 2)
	require.Equal(t, "admin_can_update_and_view_pages", policies[0].ID)
	require.Equal(t, "guest_can_only_view_pages", policies[1].ID)

	require.Equal(t, []string{"VIEW", "UPDATE"}, policies[0].Actions)
	require.Equal(t, []string{"VIEW"}, policies[1].Actions)

	require.NotEmpty(t, policies[0].Condition)
	require.NotEmpty(t, policies[1].Condition)
	require.NotEqual(t, policies[0].Condition, policies[1].Condition)

	policies, err = httpserver.LoadPoliciesFromYaml(bytes.NewBufferString(`
  - id: admin_can_update_and_view_pages
     a
  a
`))
	require.Error(t, err)
	require.Empty(t, policies)
}

func TestUnmarshalAttributes(t *testing.T) {
	testCases := []struct {
		yaml          string
		expectedAttrs leges.Attributes
		expectedErr   error
	}{
		{
			yaml:          `{}`,
			expectedAttrs: leges.Attributes{},
			expectedErr:   nil,
		},
		{
			yaml:          `{"key1": "val1"}`,
			expectedAttrs: leges.Attributes{"key1": "val1"},
			expectedErr:   nil,
		},
		{
			yaml:          `{"key1": "val1`,
			expectedAttrs: nil,
			expectedErr:   errors.New("unexpected EOF"),
		},
	}
	for _, tt := range testCases {
		attrs, err := httpserver.UnmarshalAttributes(tt.yaml)
		require.Equal(t, tt.expectedErr, err)
		require.Equal(t, tt.expectedAttrs, attrs)
	}
}

func TestMustMarshal(t *testing.T) {
	testCases := []struct {
		attrs        leges.Attributes
		expectedYaml string
		expectedErr  error
	}{
		{
			attrs:        leges.Attributes{},
			expectedYaml: `{}` + "\n",
			expectedErr:  nil,
		},
		{
			attrs:        leges.Attributes{"key1": "val1"},
			expectedYaml: `{"key1":"val1"}` + "\n",
			expectedErr:  nil,
		},
	}
	for _, tt := range testCases {
		y := httpserver.MustMarshal(tt.attrs)
		require.Equal(t, tt.expectedYaml, y)
	}
}
