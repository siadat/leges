package leges_test

import (
	"github.com/siadat/leges"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request leges.Request
		err     error
	}{
		{
			name: "valid request",
			request: leges.Request{
				Action: "UPDATE",
				Subject: leges.Attributes{
					"id": "user1",
				},
				Object: leges.Attributes{
					"type":     "account",
					"owner_id": "user1",
				},
			},
			err: nil,
		},
		{
			name: "empty object - invalid request",
			request: leges.Request{
				Action: "UPDATE",
				Subject: leges.Attributes{
					"id": "user1",
				},
				Object: leges.Attributes{},
			},
			err: leges.ErrEmptyObjectAttrs,
		},
		{
			name: "empty subject - invalid request",
			request: leges.Request{
				Action:  "UPDATE",
				Subject: leges.Attributes{},
				Object: leges.Attributes{
					"type":     "account",
					"owner_id": "user1",
				},
			},
			err: leges.ErrEmptySubjectAttrs,
		},
		{
			name: "empty action - invalid request",
			request: leges.Request{
				Action: "",
				Subject: leges.Attributes{
					"id": "user1",
				},
				Object: leges.Attributes{
					"type":     "account",
					"owner_id": "user1",
				},
			},
			err: leges.ErrEmptyAction,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			require.Equal(t, err, test.err)
		})
	}
}
