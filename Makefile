KIND ?= $(HOME)/go/bin/kind
CLUSTER ?= retry-paper
KIND_IMAGE ?= kindest/node:v1.31.4
TEKTON_VERSION ?= v1.6.0
KUBECTL ?= kubectl

.PHONY: test evaluate generate cluster install-tekton wait-tekton integration benchmark teardown

test:
	go test ./...
	go vet ./...

evaluate:
	go run ./cmd/retryctl evaluate --cases evaluation/cases.yaml

generate:
	go run ./cmd/retryctl create --pipeline testdata/simple/pipeline.yaml --run testdata/simple/run.yaml --policy testdata/simple/policy.yaml --taskruns testdata/simple/taskruns.yaml

cluster:
	$(KIND) create cluster --name $(CLUSTER) --image $(KIND_IMAGE) --wait 5m

install-tekton:
	$(KUBECTL) --context kind-$(CLUSTER) apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/$(TEKTON_VERSION)/release.yaml

wait-tekton:
	$(KUBECTL) --context kind-$(CLUSTER) wait --for=condition=Available --timeout=5m deployment --all -n tekton-pipelines

integration:
	CLUSTER=$(CLUSTER) ./hack/integration.sh

benchmark:
	CLUSTER=$(CLUSTER) ./hack/benchmark.sh

teardown:
	$(KIND) delete cluster --name $(CLUSTER)
