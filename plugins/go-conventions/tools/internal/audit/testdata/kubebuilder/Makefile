build: manifests generate fmt vet
	go build -o bin/manager cmd/main.go

manifests:
	controller-gen rbac:roleName=manager-role crd paths="./..."

generate:
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt:
	go fmt ./...

vet:
	go vet ./...

test: manifests generate fmt vet
	go test ./... -coverprofile cover.out

lint:
	golangci-lint run
