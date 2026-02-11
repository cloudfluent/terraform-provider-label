BINARY := terraform-provider-label

.PHONY: build test testacc clean fmt lint docs

build:
	go build -o $(BINARY)

test:
	TF_ACC_PROVIDER_NAMESPACE=cloudfluent go test ./... -v

testacc:
	TF_ACC=1 TF_ACC_PROVIDER_NAMESPACE=cloudfluent go test ./... -v -timeout 120s

clean:
	rm -f $(BINARY)

fmt:
	go fmt ./...
	terraform fmt -recursive examples/

lint: fmt
	go vet ./...

docs: build
	tfplugindocs generate
