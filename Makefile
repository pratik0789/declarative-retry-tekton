KIND ?= $(HOME)/go/bin/kind
CLUSTER ?= retry-paper
KIND_IMAGE ?= kindest/node:v1.31.4
TEKTON_VERSION ?= v1.6.0
KUBECTL ?= kubectl

.PHONY: test evaluate replay generate cluster install-tekton wait-tekton integration stateful ordering benchmark closure-ablation closure-smoke teardown

test:
	go test ./...
	go vet ./...

evaluate:
	go run ./cmd/retryctl evaluate --cases evaluation/cases.yaml

replay:
	for policy in policy failed-task-only; do for binding in persistentVolumeClaim volumeClaimTemplate; do \
		go run ./cmd/retryctl replay --pipeline evaluation/konflux/template-build.yaml --policy evaluation/konflux/$$policy.yaml \
			--binding $$binding --secret-workspaces git-auth,netrc > evaluation/konflux/replay-$$policy-$$binding.json; \
		go run ./cmd/retryctl replay --pipeline evaluation/konflux/template-build.yaml --policy evaluation/konflux/$$policy.yaml \
			--binding $$binding --completed independent --secret-workspaces git-auth,netrc > evaluation/konflux/replay-independent-$$policy-$$binding.json; \
	done; done

generate:
	go run ./cmd/retryctl create --pipeline testdata/simple/pipeline.yaml --run testdata/simple/run.yaml --policy testdata/simple/policy.yaml --taskruns testdata/simple/taskruns.yaml

cluster:
	$(KIND) create cluster --name $(CLUSTER) --image $(KIND_IMAGE) --wait 5m

install-tekton:
	$(KUBECTL) --context kind-$(CLUSTER) apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/$(TEKTON_VERSION)/release.yaml

wait-tekton:
	$(KUBECTL) --context kind-$(CLUSTER) wait --for=condition=Available --timeout=5m deployment --all -n tekton-pipelines
	python3 hack/wait_tekton_webhook.py --context kind-$(CLUSTER) --kubectl $(KUBECTL)

integration:
	CLUSTER=$(CLUSTER) ./hack/integration.sh
ifeq ($(CI),true)
	$(MAKE) closure-smoke
endif

closure-smoke:
	@mkdir -p artifacts
	go build -trimpath -buildvcs=false -o artifacts/retryctl-closure ./cmd/retryctl
	python3 hack/closure_ablation.py --context kind-$(CLUSTER) --namespace retry-closure-ci \
		--repetitions 1 --retryctl artifacts/retryctl-closure --output artifacts/integration/closure-ablation
	python3 hack/verify_closure_ablation.py --results artifacts/integration/closure-ablation --retryctl artifacts/retryctl-closure

stateful:
	CLUSTER=$(CLUSTER) ./hack/stateful.sh

ordering:
	CLUSTER=$(CLUSTER) ./hack/ordering.sh

benchmark:
	CLUSTER=$(CLUSTER) ./hack/benchmark.sh

# A namespace must be supplied; this experiment never resets existing namespaces.
closure-ablation:
	@test -n "$(NAMESPACE)" || (echo 'Set NAMESPACE to a new namespace name' >&2; exit 1)
	@mkdir -p artifacts
	go build -trimpath -buildvcs=false -o artifacts/retryctl-closure ./cmd/retryctl
	python3 hack/closure_ablation.py --context kind-$(CLUSTER) --namespace $(NAMESPACE) \
		--retryctl artifacts/retryctl-closure --output artifacts/closure-ablation
	python3 hack/verify_closure_ablation.py --results artifacts/closure-ablation --retryctl artifacts/retryctl-closure

teardown:
	$(KIND) delete cluster --name $(CLUSTER)
