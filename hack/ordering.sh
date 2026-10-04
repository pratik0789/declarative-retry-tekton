#!/usr/bin/env bash
# Recovery reruns build and scan (selected), inherits migrate, and reruns the failed deploy.
# deploy must still wait for the rerun scan even though its only edge pointed at migrate.
set -euo pipefail

cluster="${CLUSTER:-retry-paper}"
context="kind-${cluster}"
namespace="retry-ordering"
artifact_dir="${ARTIFACT_DIR:-artifacts/ordering}"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT
mkdir -p "${artifact_dir}"
k() { kubectl --context "${context}" -n "${namespace}" "$@"; }

kubectl --context "${context}" delete namespace "${namespace}" --ignore-not-found --wait=true
kubectl --context "${context}" create namespace "${namespace}"
k apply -f integration/ordering.yaml
k wait --for=condition=Succeeded=False --timeout=5m pipelinerun/retry-ordering-source
k get pipeline/retry-ordering -o yaml > "${tmp_dir}/pipeline.yaml"
k get pipelinerun/retry-ordering-source -o yaml > "${tmp_dir}/run.yaml"
k get taskrun -l tekton.dev/pipelineRun=retry-ordering-source -o yaml > "${tmp_dir}/taskruns.yaml"

go run ./cmd/retryctl create --pipeline "${tmp_dir}/pipeline.yaml" --run "${tmp_dir}/run.yaml" \
  --policy integration/ordering-policy.yaml --taskruns "${tmp_dir}/taskruns.yaml" --tasks scan \
  > "${artifact_dir}/recovery.yaml"
recovery="$(k create -f "${artifact_dir}/recovery.yaml" -o name)"
k wait --for=condition=Succeeded=True --timeout=5m "${recovery}"
recovery="${recovery#*/}"
k get taskrun -l "tekton.dev/pipelineRun=${recovery}" -o json > "${artifact_dir}/recovery-taskruns.json"

python3 - "${artifact_dir}/recovery-taskruns.json" "${artifact_dir}/summary.json" <<'PY'
import json, sys
from datetime import datetime
items = json.load(open(sys.argv[1]))["items"]
t = lambda s: datetime.fromisoformat(s.replace("Z", "+00:00"))
runs = {i["metadata"]["labels"]["tekton.dev/pipelineTask"]: i["status"] for i in items}
summary = {
    "tasksRun": sorted(runs),
    "scanCompleted": runs["scan"]["completionTime"],
    "deployStarted": runs["deploy"]["startTime"],
    "deployStartedAfterScan": t(runs["deploy"]["startTime"]) >= t(runs["scan"]["completionTime"]),
}
json.dump(summary, open(sys.argv[2], "w"), indent=2)
print(json.dumps(summary, indent=2))
PY
