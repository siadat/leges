# 👩🏻‍⚖️ Leges

Leges is an attribute-based access-control HTTP service and Go library.

[![GoDoc](https://godoc.org/github.com/siadat/leges?status.svg)](https://godoc.org/github.com/siadat/leges)
[![Build Status](https://travis-ci.org/siadat/leges.svg?branch=master)](https://travis-ci.org/siadat/leges)
[![Coverage Status](https://codecov.io/gh/siadat/leges/branch/master/graph/badge.svg)](https://codecov.io/gh/siadat/leges)
[![Go Report Card](https://goreportcard.com/badge/github.com/siadat/leges)](https://goreportcard.com/report/github.com/siadat/leges)

## Examples

* Let thirsty cats and elephants drink water
```
condition: |
  subject.kind in ["cat", "elephant"]
  and subject.thirsty
  and object.name == "water"

actions: [DRINK]
```

* Let users view and edit their own post
```
condition: |
  subject.id == object.owner_user_id

actions: [VIEW_POST, EDIT_POST]
```



## HTTP service

Describe the policies in a YAML file:

```yaml
- id: admin_can_update_and_view_pages_and_adminpages
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
```

Start the leges service:

```bash
leges --addr :5120 --policies sample-policies.yaml
```


Send a request to see if a guest can VIEW a page:

```
# Pseudo-HTTP request
GET /match?object  = {"role": "guest"}
          &subject = {"type": "page"}
          &action  = VIEW
```

Using curl (same request):
```bash
$ curl 'http://localhost:5120/match?object=%7B%22type%22%3A%22page%22%7D&subject=%7B%22role%22%3A%22guest%22%7D&action=VIEW'
{
  "match": true,
  "id": "guest_can_only_view_pages"
}
```

**Note:** subject and object should be given as URI-encoded JSON values. For example, In Javascript 
[encodeURIComponent](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/encodeURIComponent)
and in Python3 [urllib.parse.quote](https://docs.python.org/3/library/urllib.parse.html#urllib.parse.quote)
should be used.

## Go library

Build your own HTTP/gRPC/etc service using the Go library described below.

We will implement the following policies:

- admin users can view and update a page and an adminpage
- guest users can only view a page

First, create an instance of leges.Leges:

```go
lg, err := leges.NewLeges([]leges.Policy{
	{
		// ID is used to identify the matching policy
		ID: "admin_can_update_and_view_pages",

		// Condition describe the policy using an expression
		Condition: `
			subject.role == "admin"
			and object.type in ["page", "adminpage"]
		`,

		// Actions is a list of actions allowed by this policy
		Actions: []string{
			"VIEW",
			"UPDATE",
		},
	},
	{
		// ID is used to identify the matching policy
		ID: "guest_can_only_view_pages",

		// Condition describe the policy using an expression
		Condition: `
			subject.role == "guest"
			and object.type == "page"
		`,

		// Actions is a list of actions allowed by this policy
		Actions: []string{
			"VIEW",
		},
	},
}, nil)
```

Let's say a request arrives to update a page by a user whose role is "guest":

```go
request := leges.Request{
	Action: "UPDATE",
	Subject: leges.Attributes{
		"role": "guest",
	},
	Object: leges.Attributes{
		"type": "page",
	},
}

ok, policy, err := lg.Match(request)
// ok:     false
// policy: nil
// err:    nil
```

No policy exists for a guest to update a page (only admins can do that), so leges.Match returns false.

## Trivia

Leges is Latin for *laws*. We define the laws (via lg.NewLeges)
and ask leges (via lg.Match) to judge whether a given request
is allowed or not.
