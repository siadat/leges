package leges_test

import (
	"github.com/siadat/leges"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPolicy_Validate(t *testing.T) {
	t.Run("id empty - invalid policy", func(t *testing.T) {
		p := leges.Policy{
			ID:        "",
			Condition: "subject.is_admin == true",
			Actions:   []string{"action"},
		}

		err := p.Validate()
		require.Error(t, err, leges.ErrEmptyPolicyID)
	})

	t.Run("valid policy", func(t *testing.T) {
		p := leges.Policy{
			ID:        "id",
			Condition: "subject.is_admin == true",
			Actions:   []string{"action"},
		}

		err := p.Validate()
		require.NoError(t, err)
	})
}
