test:
	GO111MODULE=on go vet ./...
	GO111MODULE=on go test -i .
	GO111MODULE=on go test -v -failfast -race -count=1 -coverpkg=./... -coverprofile coverage_tmp.out -covermode=atomic ./...
	@cat coverage_tmp.out | sed -n '1p' > coverage.out
	@cat coverage_tmp.out | sed -n '2,$$p' >> coverage.out
	@rm coverage_tmp.out
	GO111MODULE=on go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage: file://$(PWD)/coverage.html"

benchmark:
	GO111MODULE=on go test -bench=Match -benchtime=2s
