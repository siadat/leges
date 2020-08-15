package leges

import (
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"sync"
)

var ErrDuplicatePolicyID = errors.New("duplicate policies with id")

// Leges is the law book, holding all policies and the base environment
type Leges struct {
	// cachedPolicies is a list of cachedPolicy, making the law
	cachedPolicies map[string]cachedPolicy
	// environment are a set of attributes that are always merged with the request
	environment Attributes

	// lock provides mutex for cachedPolicies
	lock sync.Mutex
}

type ErrExprRunFailed struct {
	Environment Attributes
	Policy      Policy
	Request     Request
	Err         error
}

func (e *ErrExprRunFailed) Unwrap() error {
	return e.Err
}

func (e *ErrExprRunFailed) Error() string {
	return "failed to run expression"
}

type ErrExprCompileFailed struct {
	Environment Attributes
	Policy      Policy
	Err         error
}

func (e *ErrExprCompileFailed) Unwrap() error {
	return e.Err
}

func (e *ErrExprCompileFailed) Error() string {
	return "failed to compile expression"
}

// cachedPolicy is one policy and the compiled program of the condition
type cachedPolicy struct {
	policy  Policy
	program *vm.Program
}

// Attributes is a set of key-value attributes for objects and subjects.
type Attributes = map[string]interface{}

// NewLeges construct a leges struct
func NewLeges(policies []Policy, env Attributes) (*Leges, error) {
	leges := &Leges{}

	if err := leges.loadEnvironment(env); err != nil {
		return nil, err
	}

	if err := leges.loadPolicies(policies); err != nil {
		return nil, err
	}

	return leges, nil
}

func (l *Leges) loadEnvironment(env Attributes) error {
	l.environment = Attributes{}

	for k, v := range env {
		l.environment[k] = v
	}

	return nil
}

// loadPolicies iterates over a list of policies, runs validation on each of them and
// keeps a compiled program of the condition for further use cases
func (l *Leges) loadPolicies(polices []Policy) error {
	l.cachedPolicies = make(map[string]cachedPolicy, len(polices))

	for _, policy := range polices {
		if err := l.AddPolicy(policy); err != nil {
			return err
		}
	}

	return nil
}

func (l *Leges) AddPolicy(policy Policy) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	if err := policy.Validate(); err != nil {
		return err
	}

	if _, ok := l.cachedPolicies[policy.ID]; ok {
		return fmt.Errorf("id=%q: %w", policy.ID, ErrDuplicatePolicyID)
	}

	program, err := policy.compileCondition()
	if err != nil {
		return &ErrExprCompileFailed{
			Environment: l.environment,
			Policy:      policy,
			Err:         err,
		}
	}

	l.cachedPolicies[policy.ID] = cachedPolicy{
		policy:  policy,
		program: program,
	}

	return nil
}

// normalizeRequest normalizes the request by merging it with environment
func (l *Leges) normalizeRequest(request Request) Attributes {
	req := Attributes{}

	// copy all environment to req
	for k, v := range l.environment {
		req[k] = v
	}

	// merge the request
	req["subject"] = request.Subject
	req["object"] = request.Object

	req["debug"] = func(value interface{}) bool {
		fmt.Printf("%#v\n", value)
		return true
	}

	return req
}

// Match checks a request against policies and returns whether the request matches any of the policies.
func (l *Leges) Match(request Request) (bool, *Policy, error) {
	if err := request.Validate(); err != nil {
		return false, nil, err
	}

	normalizedRequest := l.normalizeRequest(request)

	for _, statute := range l.cachedPolicies {
		if !sliceIncludes(statute.policy.Actions, request.Action) {
			continue
		}

		output, err := expr.Run(statute.program, normalizedRequest)
		if err != nil {
			return false, nil, &ErrExprRunFailed{
				Err:         err,
				Environment: l.environment,
				Policy:      statute.policy,
				Request:     request,
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
