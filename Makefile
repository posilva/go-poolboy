
PHONY: test
test:
	go test -timeout 20s -v -race -parallel 4 -count 1 -cover -coverprofile=coverage.out && go tool cover -html=coverage.out
vet: 
	go vet