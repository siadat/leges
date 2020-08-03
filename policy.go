package leges

import (
	"errors"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

var ErrEmptyPolicyID = errors.New("policy with empty id")

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

func (p Policy) Validate() error {
	if p.ID == "" {
		return ErrEmptyPolicyID
	}
	return nil
}

func (p Policy) compileCondition() (*vm.Program, error) {
	return expr.Compile(p.Condition)
}
