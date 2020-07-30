package leges

import (
	"errors"
	"fmt"
	"sync"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

var conditionProgramCache = struct {
	lock  sync.RWMutex
	items map[string]*vm.Program
}{
	lock:  sync.RWMutex{},
	items: map[string]*vm.Program{},
}

// Attributes is a set of key-value attributes for objects and subjects.
type Attributes = map[string]interface{}

// Policy defines a condition that is allowed.
type Policy struct {
	// ID is used to identify the policy when matching
	ID string
	// Condition specifies an expression that should be true for the policy
	// to match. For example, "subject.is_admin == true".
	Condition string
	// Actions is a list of actions allowed for this policy. For example,
	// []string{"GET", "SET"} means that this policy allows both GET and
	// SET actions.
	Actions []string
}

// Request defines a request to be checked against the policies.
type Request struct {
	// Action is the requested action name.
	Action string
	// Subject is the attributes of the subject/requester.
	Subject Attributes
	// Object is the attributes of the object/resource.
	Object Attributes
}

func sliceIncludes(slice []string, needle string) bool {
	for _, item := range slice {
		if item == needle {
			return true
		}
	}
	return false
}

var (
	ErrEmptyPolicyID     = errors.New("policy with empty id")
	ErrEmptyObjectAttrs  = errors.New("object attributes is empty")
	ErrEmptySubjectAttrs = errors.New("subject attributes is empty")
	ErrEmptyAction       = errors.New("action is empty")
	ErrDuplicatePolicyID = errors.New("duplicate policies with id")
)

type ErrExprRunFailed struct {
	Env     Attributes
	Policy  Policy
	Request Request
	Err     error
}

func (e *ErrExprRunFailed) Unwrap() error {
	return e.Err
}

func (e *ErrExprRunFailed) Error() string {
	return "failed to run expression"
}

type ErrExprCompileFailed struct {
	Env     Attributes
	Policy  Policy
	Request Request
	Err     error
}

func (e *ErrExprCompileFailed) Unwrap() error {
	return e.Err
}

func (e *ErrExprCompileFailed) Error() string {
	return "failed to compile expression"
}

// Leges is the struct that contains the policies and can be used to requests
type Leges struct {
	policies []Policy
}

func NewLeges(policies []Policy) (*Leges, error) {
	visitedPolicyIDs := map[string]struct{}{}
	for _, p := range policies {
		if p.ID == "" {
			return nil, ErrEmptyPolicyID
		}
		if _, ok := visitedPolicyIDs[p.ID]; ok {
			return nil, fmt.Errorf("id=%q: %w", p.ID, ErrDuplicatePolicyID)
		}
		visitedPolicyIDs[p.ID] = struct{}{}
	}

	return &Leges{
		policies: policies,
	}, nil
}

// Match checks a request against policies and returns whether the request matches any of the policies.
func (lg *Leges) Match(request Request, userEnv Attributes) (bool, *Policy, error) {
	if request.Object == nil || len(request.Object) == 0 {
		return false, nil, ErrEmptyObjectAttrs
	}
	if request.Subject == nil || len(request.Subject) == 0 {
		return false, nil, ErrEmptySubjectAttrs
	}
	if request.Action == "" {
		return false, nil, ErrEmptyAction
	}

	env := map[string]interface{}{}

	// we need to copy userEnv, because we want to write to userEnv before
	// evaluating it ("subject" and "object" keys), and the received
	// userEnv map might be written to elsewhere by the caller.
	for k, v := range userEnv {
		env[k] = v
	}

	// We do this conversion in order to be able to write conditions like:
	//   object == {}
	//   object == {key1: "val1", key2: "val2"}
	// Without this conversion, the types will not match and we will have to write:
	//   len(object) == 0
	//   object.key1 == "val1" and object.key2 == "val2"
	env["subject"] = map[string]interface{}(request.Subject)
	env["object"] = map[string]interface{}(request.Object)

	env["debug"] = func(value interface{}) bool {
		fmt.Printf("%#v\n", value)
		return true
	}

	for _, policy := range lg.policies {
		if !sliceIncludes(policy.Actions, request.Action) {
			continue
		}

		conditionProgramCache.lock.RLock()
		program, ok := conditionProgramCache.items[policy.Condition]
		conditionProgramCache.lock.RUnlock()

		if !ok {
			var err error
			program, err = expr.Compile(policy.Condition)
			if err != nil {
				return false, nil, &ErrExprCompileFailed{
					Err:     err,
					Env:     env,
					Policy:  policy,
					Request: request,
				}
			}
			conditionProgramCache.lock.Lock()
			conditionProgramCache.items[policy.Condition] = program
			conditionProgramCache.lock.Unlock()
		}

		output, err := expr.Run(program, env)
		if err != nil {
			return false, nil, &ErrExprRunFailed{
				Err:     err,
				Env:     env,
				Policy:  policy,
				Request: request,
			}
		}

		if output.(bool) {
			return true, &policy, nil
		}
	}

	return false, nil, nil
}
