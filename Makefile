test:
	GO111MODULE=on go vet ./...
	GO111MODULE=on go test -i .
	GO111MODULE=on go test -v -failfast -race -count=1 -coverpkg=./... -coverprofile coverage.out -covermode=atomic ./...
	GO111MODULE=on go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage: file://$(PWD)/coverage.html"

benchmark:
	GO111MODULE=on go test -bench=Match -benchtime=2s
