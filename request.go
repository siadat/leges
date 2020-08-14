package leges

import "errors"

var (
	ErrEmptyObjectAttrs  = errors.New("object attributes is empty")
	ErrEmptySubjectAttrs = errors.New("subject attributes is empty")
	ErrEmptyAction       = errors.New("action is empty")
)

// Request defines a request to be checked against the policies.
type Request struct {
	// Action is the requested action name.
	Action string
	// Subject is the attributes of the subject/requester.
	Subject Attributes
	// Object is the attributes of the object/resource.
	Object Attributes
}

func (r Request) Validate() error {
	if r.Object == nil || len(r.Object) == 0 {
		return ErrEmptyObjectAttrs
	}
	if r.Subject == nil || len(r.Subject) == 0 {
		return ErrEmptySubjectAttrs
	}
	if r.Action == "" {
		return ErrEmptyAction
	}
	return nil
}
