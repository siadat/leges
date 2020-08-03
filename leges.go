package leges

import (
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

var ErrDuplicatePolicyID = errors.New("duplicate policies with id")

// Leges is the law book, holding all policies and the base definitions
type Leges struct {
	// statutes is a list of statute, making the law
	statutes map[string]statute
	// definitions are a set of attributes that are always merged with the request to form a motion
	definitions Attributes
}

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
	Env    Attributes
	Policy Policy
	Err    error
}

func (e *ErrExprCompileFailed) Unwrap() error {
	return e.Err
}

func (e *ErrExprCompileFailed) Error() string {
	return "failed to compileCondition expression"
}

// statute is one policy and the compiled program of the condition
type statute struct {
	policy  Policy
	program *vm.Program
}

// Attributes is a set of key-value attributes for objects and subjects.
type Attributes = map[string]interface{}

func New(policies []Policy, definitions Attributes) (*Leges, error) {
	leges := &Leges{}

	if err := leges.loadDefinitions(definitions); err != nil {
		return nil, err
	}

	if err := leges.loadPolicies(policies); err != nil {
		return nil, err
	}

	return leges, nil
}

func (l *Leges) loadDefinitions(definitions Attributes) error {
	l.definitions = Attributes{}

	for k, v := range definitions {
		l.definitions[k] = v
	}

	return nil
}

// loadPolicies iterates over a list of policies, runs validation on each of them and
// keeps a compiled program of the condition for further use cases
func (l *Leges) loadPolicies(polices []Policy) error {
	l.statutes = make(map[string]statute, len(polices))

	for _, policy := range polices {
		if err := policy.Validate(); err != nil {
			return err
		}

		if _, ok := l.statutes[policy.ID]; ok {
			return fmt.Errorf("id=%q: %w", policy.ID, ErrDuplicatePolicyID)
		}

		program, err := policy.compileCondition()
		if err != nil {
			return &ErrExprCompileFailed{
				Env:    l.definitions,
				Policy: policy,
				Err:    err,
			}
		}

		l.statutes[policy.ID] = statute{
			policy:  policy,
			program: program,
		}
	}

	return nil
}

// makeClaim merges the request with the definitions to form the claim
func (l *Leges) makeClaim(request Request) Attributes {
	claim := Attributes{}

	// we need to copy userEnv, because we want to write to userEnv before
	// evaluating it ("subject" and "object" keys), and the received
	// userEnv map might be written to elsewhere by the caller.
	for k, v := range l.definitions {
		claim[k] = v
	}

	claim["subject"] = request.Subject
	claim["object"] = request.Object

	claim["debug"] = func(value interface{}) bool {
		fmt.Printf("%#v\n", value)
		return true
	}

	return claim
}

// Match checks a request against policies and returns whether the request matches any of the policies.
func (l *Leges) Match(request Request) (bool, *Policy, error) {
	if err := request.Validate(); err != nil {
		return false, nil, err
	}

	claim := l.makeClaim(request)

	for _, statute := range l.statutes {
		if !sliceIncludes(statute.policy.Actions, request.Action) {
			continue
		}

		output, err := expr.Run(statute.program, claim)
		if err != nil {
			return false, nil, &ErrExprRunFailed{
				Err:     err,
				Env:     claim,
				Policy:  statute.policy,
				Request: request,
			}
		}

		if output.(bool) {
			return true, &statute.policy, nil
		}
	}

	return false, nil, nil
}

func sliceIncludes(slice []string, needle string) bool {
	for _, item := range slice {
		if item == needle {
			return true
		}
	}
	return false
}
