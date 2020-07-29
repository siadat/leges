package leges_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/siadat/leges"
	"github.com/stretchr/testify/require"
)

func ExampleMatch() {
	policies := []leges.Policy{
		{
			ID: "user1_can_view_session",
			Condition: `
				object.type == "session"
				and subject.id == "user1"
				and key1 == "value1"
			`,
			Actions: []string{
				"VIEW",
			},
		},
	}

	sharedEnv := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	ok, policy, err := leges.Match(policies, leges.Request{
		Action: "VIEW",
		Subject: leges.Attributes{
			"id": "user1",
		},
		Object: leges.Attributes{
			"type": "session",
		},
	}, sharedEnv)

	fmt.Printf("ok: %v\n", ok)
	fmt.Printf("policy.ID: %q\n", policy.ID)
	fmt.Printf("err: %v\n", err)
	// Output:
	// ok: true
	// policy.ID: "user1_can_view_session"
	// err: <nil>
}

func BenchmarkMatch(b *testing.B) {
	policies := []leges.Policy{
		{
			ID: "policy1",
			Condition: `
				object.type == "session"
				and subject.id == "anonymous"
			`,
			Actions: []string{
				"SIGNUP",
				"LOGIN",
				"REFRESH",
			},
		},
		{
			ID: "policy2",
			Condition: `
				object.type == "account"
				and object.owner_id == subject.id
			`,
			Actions: []string{
				"PUBLIC_VIEW",
				"GET",
				"UPDATE",
				"MY_ACTION",
			},
		},
		{
			ID: "policy3",
			Condition: `
				object.attr1 == "val"
				and object.attr2 == subject.attr1
				and object.attr3 == subject.attr2
			`,
			Actions: []string{
				"MY_ACTION",
			},
		},
		{
			ID: "policy4",
			Condition: `
				object.attr1 == "val"
				and object.attr2 == subject.attr1
			`,
			Actions: []string{
				"MY_ACTION",
			},
		},
		{
			ID: "policy5",
			Condition: `
				subject.id in object.shared_with
			`,
			Actions: []string{
				"SEE",
			},
		},
	}
	sharedEnv := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	var (
		ok     bool
		policy *leges.Policy
		err    error
	)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ok, policy, err = leges.Match(policies, leges.Request{
				Action: "UPDATE",
				Subject: leges.Attributes{
					"id": "user1",
				},
				Object: leges.Attributes{
					"type":     "account",
					"owner_id": "user1",
				},
			}, sharedEnv)
		}
	})
	benchOk = ok
	benchPolicy = policy
	benchErr = err

}

var (
	benchOk     bool
	benchPolicy *leges.Policy
	benchErr    error
)

func TestSimplePolicy(t *testing.T) {
	var policies []leges.Policy

	t.Run("error if subject is empty", func(t *testing.T) {
		_, _, err := leges.Match(policies, leges.Request{
			Action:  "ACTION",
			Subject: leges.Attributes{},
			Object: leges.Attributes{
				"att1": "attr1",
			},
		}, nil)
		require.Error(t, err)
		require.Equal(t, leges.ErrEmptySubjectAttrs, err)
	})

	t.Run("error if object is empty", func(t *testing.T) {
		_, _, err := leges.Match(policies, leges.Request{
			Action: "ACTION",
			Subject: leges.Attributes{
				"att1": "attr1",
			},
			Object: leges.Attributes{},
		}, nil)
		fmt.Printf("err = %+v\n", err)
		require.Error(t, err)
		require.Equal(t, leges.ErrEmptyObjectAttrs, err)
	})

	t.Run("no policies", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "UPDATE",
			Subject: leges.Attributes{
				"id": "user1",
			},
			Object: leges.Attributes{
				"type":     "account",
				"owner_id": "user1",
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, false, ok)
		require.Nil(t, policy)
	})

	policies = []leges.Policy{
		{
			ID: "id1",
			Condition: `
				object.type == session
			`,
			Actions: []string{
				"ACTION",
			},
		},
		{
			ID: "id1",
			Condition: `
				object.type != session
			`,
			Actions: []string{
				"ACTION",
			},
		},
	}

	t.Run("err if policies have duplicate IDs", func(t *testing.T) {
		_, _, err := leges.Match(policies, leges.Request{
			Action: "ACTION",
			Subject: leges.Attributes{
				"att1": "attr1",
			},
			Object: leges.Attributes{
				"att1": "attr1",
			},
		}, nil)
		require.Error(t, err)
		require.True(t, errors.Is(err, leges.ErrDuplicatePolicyID))
	})

	policies = []leges.Policy{
		{
			ID: "",
			Condition: `
				object.type == session
			`,
			Actions: []string{
				"ACTION",
			},
		},
	}

	t.Run("err if policy has no id", func(t *testing.T) {
		_, _, err := leges.Match(policies, leges.Request{
			Action: "ACTION",
			Subject: leges.Attributes{
				"att1": "attr1",
			},
			Object: leges.Attributes{
				"att1": "attr1",
			},
		}, nil)
		require.Error(t, err)
		require.Equal(t, leges.ErrEmptyPolicyID, err)
	})

	policies = []leges.Policy{
		{
			ID: "policy1",
			Condition: `
				object.type == session"
				and subject.id == "anonymous"
			`,
			Actions: []string{
				"ACTION",
			},
		},
	}

	t.Run("error if policy cant be compiled", func(t *testing.T) {
		_, _, err := leges.Match(policies, leges.Request{
			Action: "ACTION",
			Subject: leges.Attributes{
				"att1": "attr1",
			},
			Object: leges.Attributes{
				"att1": "attr1",
			},
		}, nil)
		require.Error(t, err)
		require.IsType(t, &leges.ErrExprCompileFailed{}, err)
	})

	policies = []leges.Policy{
		{
			ID: "policy1",
			Condition: `
				object.type == "session"
				and subject.id == "anonymous"
			`,
			Actions: []string{
				"SIGNUP",
				"LOGIN",
				"REFRESH",
			},
		},
		{
			ID: "policy2",
			Condition: `
				object.type == "account"
				and object.owner_id == subject.id
			`,
			Actions: []string{
				"PUBLIC_VIEW",
				"GET",
				"UPDATE",
				"MY_ACTION",
			},
		},
		{
			ID: "policy3",
			Condition: `
				object.attr1 == "val"
				and object.attr2 == subject.attr1
				and object.attr3 == subject.attr2
			`,
			Actions: []string{
				"MY_ACTION",
			},
		},
		{
			ID: "policy4",
			Condition: `
				object.attr1 == "val"
				and object.attr2 == subject.attr1
			`,
			Actions: []string{
				"MY_ACTION",
			},
		},
		{
			ID: "policy5",
			Condition: `
				subject.id in object.shared_with
			`,
			Actions: []string{
				"SEE",
			},
		},
	}

	t.Run("match $subject.id and $object.owner_id", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "UPDATE",
			Subject: leges.Attributes{
				"id": "user1",
			},
			Object: leges.Attributes{
				"type":     "account",
				"owner_id": "user1",
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, true, ok)
		require.Equal(t, "policy2", policy.ID)
	})

	t.Run("match something with env while doing it concurrently to check for race conditions", func(t *testing.T) {
		sharedEnv := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		wg := sync.WaitGroup{}
		for i := 0; i < 10000; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				ok, policy, err := leges.Match(policies, leges.Request{
					Action: "UPDATE",
					Subject: leges.Attributes{
						"id": "user1",
					},
					Object: leges.Attributes{
						"type":     "account",
						"owner_id": "user1",
					},
				}, sharedEnv)
				require.NoError(t, err)
				require.Equal(t, true, ok)
				require.Equal(t, "policy2", policy.ID)
			}(i)
		}
		wg.Wait()
	})

	t.Run("allow anonymous to signup", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "SIGNUP",
			Subject: leges.Attributes{
				"id": "anonymous",
			},
			Object: leges.Attributes{
				"type": "session",
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, true, ok)
		require.Equal(t, "policy1", policy.ID)
	})

	t.Run("allow anonymous to login", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "LOGIN",
			Subject: leges.Attributes{
				"id": "anonymous",
			},
			Object: leges.Attributes{
				"type": "session",
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, true, ok)
		require.Equal(t, "policy1", policy.ID)
	})

	t.Run("dont allow anonymous to 'HACK'", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "HACK",
			Subject: leges.Attributes{
				"id": "anonymous",
			},
			Object: leges.Attributes{
				"type": "session",
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, false, ok)
		require.Nil(t, policy)
	})

	t.Run("dont allow not-anonymous to signup", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "SIGNUP",
			Subject: leges.Attributes{
				"id": "not-anonymous",
			},
			Object: leges.Attributes{
				"type": "session",
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, false, ok)
		require.Nil(t, policy)
	})

	t.Run("dont allow anonymous to signup not-session", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "SIGNUP",
			Subject: leges.Attributes{
				"id": "anonymous",
			},
			Object: leges.Attributes{
				"type": "not-session",
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, false, ok)
		require.Nil(t, policy)
	})

	t.Run("multiple attrbutes match", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "MY_ACTION",
			Subject: leges.Attributes{
				"attr1": "some_value_A",
				"attr2": "some_value_B",
			},
			Object: leges.Attributes{
				"attr1": "val",
				"attr2": "some_value_A",
				"attr3": "some_value_B",
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, true, ok)
		require.Equal(t, "policy3", policy.ID)
	})

	t.Run("multiple attrbutes match 2", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "MY_ACTION",
			Subject: leges.Attributes{
				"attr1": "some_value_A",
				"attr2": "some_value_B",
			},
			Object: leges.Attributes{
				"attr1": "val",
				"attr2": "some_value_A",
				"attr3": "not some_value_B",
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, true, ok)
		require.Equal(t, "policy4", policy.ID)
	})

	t.Run("match if shared_with", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "SEE",
			Subject: leges.Attributes{
				"id": "uid1",
			},
			Object: leges.Attributes{
				"shared_with": []string{"uid0", "uid1", "uid2"},
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, true, ok)
		require.Equal(t, "policy5", policy.ID)
	})

	t.Run("dont match if not shared_with", func(t *testing.T) {
		ok, policy, err := leges.Match(policies, leges.Request{
			Action: "SEE",
			Subject: leges.Attributes{
				"id": "uid100",
			},
			Object: leges.Attributes{
				"shared_with": []string{"uid0", "uid1", "uid2"},
			},
		}, nil)
		require.NoError(t, err)
		require.Equal(t, false, ok)
		require.Nil(t, policy)
	})

	policies = []leges.Policy{
		{
			ID: "policy1",
			Condition: `
				bad_expression.type == "session"
			`,
			Actions: []string{
				"ACTION",
			},
		},
	}

	t.Run("error if expression is invalid", func(t *testing.T) {
		_, _, err := leges.Match(policies, leges.Request{
			Action: "ACTION",
			Subject: leges.Attributes{
				"att1": "attr1",
			},
			Object: leges.Attributes{
				"att1": "attr1",
			},
		}, nil)
		require.Error(t, err)
		require.IsType(t, &leges.ErrExprRunFailed{}, err)
	})

}
