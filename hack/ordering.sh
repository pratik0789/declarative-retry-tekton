#!/usr/bin/env bash
# Two live ordering checks plus a cluster-sourced retryOnlyLatest check.
# Linear: build and scan rerun (selected scan), migrate is inherited, the failed deploy reruns.
# Fan-in: build-a reruns (selected), build-b, package and publish are inherited, deploy reruns.
# In both, deploy's only edge points at an inherited task, yet it must wait for the rerun task above it.
set -euo pipefail

cluster="${CLUSTER:-retry-paper}"
context="kind-${cluster}"
namespace="retry-ordering"
artifact_dir="${ARTIFACT_DIR:-artifacts/ordering}"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT
mkdir -p "${artifact_dir}"
k() { kubectl --context "${context}" -n "${namespace}" "$@"; }

go build -o "${tmp_dir}/retryctl" ./cmd/retryctl
kubectl --context "${context}" delete namespace "${namespace}" --ignore-not-found --wait=true
kubectl --context "${context}" create namespace "${namespace}"
k apply -f integration/ordering.yaml -f integration/ordering-fanin.yaml
k wait --for=condition=Succeeded=False --timeout=5m pipelinerun/retry-ordering-source pipelinerun/retry-fanin-source

export_source() {
  k get "pipeline/$1" -o yaml > "${tmp_dir}/$1-pipeline.yaml"
  k get "pipelinerun/$1-source" -o yaml > "${tmp_dir}/$1-run.yaml"
  k get taskrun -l "tekton.dev/pipelineRun=$1-source" -o yaml > "${tmp_dir}/$1-taskruns.yaml"
}
plan_decision() {
  "${tmp_dir}/retryctl" plan --pipeline "${tmp_dir}/$1-pipeline.yaml" --run "${tmp_dir}/$1-run.yaml" \
    --policy "$2" --taskruns "${tmp_dir}/$1-taskruns.yaml" --newer-runs-from-cluster --kube-context "${context}" \
    | python3 -c 'import json,sys; p=json.load(sys.stdin); print(p["decision"] + (": " + p["refusalReason"] if p.get("refusalReason") else ""))'
}
recover() {
  "${tmp_dir}/retryctl" create --pipeline "${tmp_dir}/$1-pipeline.yaml" --run "${tmp_dir}/$1-run.yaml" \
    --policy "$2" --taskruns "${tmp_dir}/$1-taskruns.yaml" --tasks "$3" --confirm-warnings "${@:4}" \
    2>/dev/null > "${artifact_dir}/$1-recovery.yaml"
  local name
  name="$(k create -f "${artifact_dir}/$1-recovery.yaml" -o name)"
  k wait --for=condition=Succeeded=True --timeout=5m "${name}" >/dev/null
  k get taskrun -l "tekton.dev/pipelineRun=${name#*/}" -o json > "${artifact_dir}/$1-recovery-taskruns.json"
}

export_source retry-ordering
recover retry-ordering integration/ordering-policy.yaml scan

export_source retry-fanin
before="$(plan_decision retry-fanin integration/ordering-fanin-policy.yaml)"
recover retry-fanin integration/ordering-fanin-policy.yaml build-a --newer-runs-from-cluster --kube-context "${context}"
after="$(plan_decision retry-fanin integration/ordering-fanin-policy.yaml)"

BEFORE="${before}" AFTER="${after}" python3 - "${artifact_dir}" <<'PY'
import json, os, sys
from datetime import datetime
d = sys.argv[1]
t = lambda s: datetime.fromisoformat(s.replace("Z", "+00:00"))
def check(name, upstream):
    runs = {i["metadata"]["labels"]["tekton.dev/pipelineTask"]: i["status"] for i in json.load(open(f"{d}/{name}-recovery-taskruns.json"))["items"]}
    return {"tasksRun": sorted(runs), "upstreamRerunCompleted": runs[upstream]["completionTime"],
            "deployStarted": runs["deploy"]["startTime"],
            "deployStartedAfterUpstream": t(runs["deploy"]["startTime"]) >= t(runs[upstream]["completionTime"])}
summary = {
    "linear": check("retry-ordering", "scan"),
    "fanIn": check("retry-fanin", "build-a"),
    "retryOnlyLatestFromCluster": {"beforeRecovery": os.environ["BEFORE"], "afterRecovery": os.environ["AFTER"]},
}
json.dump(summary, open(f"{d}/summary.json", "w"), indent=2)
print(json.dumps(summary, indent=2))
PY
