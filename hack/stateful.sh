#!/usr/bin/env bash
set -euo pipefail

cluster="${CLUSTER:-retry-paper}"
context="kind-${cluster}"
namespace="retry-stateful"
artifact_dir="${ARTIFACT_DIR:-artifacts/stateful}"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT
mkdir -p "${artifact_dir}"
k() { kubectl --context "${context}" -n "${namespace}" "$@"; }

go build -o "${tmp_dir}/retryctl" ./cmd/retryctl
kubectl --context "${context}" delete namespace "${namespace}" --ignore-not-found --wait=true
kubectl --context "${context}" create namespace "${namespace}"
k apply -f integration/stateful.yaml
k wait --for=condition=Succeeded=False --timeout=5m pipelinerun/retry-stateful-pvc-source pipelinerun/retry-stateful-vct-source
k get pipeline/retry-stateful -o yaml > "${tmp_dir}/pipeline.yaml"

export_source() {
  k get "pipelinerun/$1" -o yaml > "${tmp_dir}/$1.yaml"
  k get taskrun -l "tekton.dev/pipelineRun=$1" -o yaml > "${tmp_dir}/$1-taskruns.yaml"
}
task_field() { k get taskrun -l "tekton.dev/pipelineRun=$1,tekton.dev/pipelineTask=$2" -o "jsonpath=$3"; }
task_log() { k logs "pod/$(task_field "$1" "$2" '{.items[0].status.podName}')" --all-containers 2>/dev/null | grep -E 'STATE-' || true; }
plan_warnings() {
  "${tmp_dir}/retryctl" plan --pipeline "${tmp_dir}/pipeline.yaml" --run "${tmp_dir}/$1.yaml" \
    --policy integration/stateful-policy.yaml --taskruns "${tmp_dir}/$1-taskruns.yaml" > "${artifact_dir}/$1-plan.json"
  python3 -c 'import json,sys; print(",".join(sorted(w["code"] for w in json.load(open(sys.argv[1])).get("warnings", []))))' "${artifact_dir}/$1-plan.json"
}
recover() {
  "${tmp_dir}/retryctl" create --pipeline "${tmp_dir}/pipeline.yaml" --run "${tmp_dir}/$1.yaml" \
    --policy integration/stateful-policy.yaml --taskruns "${tmp_dir}/$1-taskruns.yaml" --confirm-warnings \
    2>/dev/null > "${artifact_dir}/$1-recovery.yaml"
  k create -f "${artifact_dir}/$1-recovery.yaml" -o name
}

# Positive case: the Workspace is a persistent claim, so the rerun build must find clone's file.
export_source retry-stateful-pvc-source
pvc_warnings="$(plan_warnings retry-stateful-pvc-source)"
pvc_recovery="$(recover retry-stateful-pvc-source)"
k wait --for=condition=Succeeded=True --timeout=5m "${pvc_recovery}"
pvc_recovery="${pvc_recovery#*/}"
source_commit="$(task_field retry-stateful-pvc-source clone '{.items[0].status.results[?(@.name=="commit")].value}')"
recovery_commit="$(task_field "${pvc_recovery}" build '{.items[0].spec.params[?(@.name=="commit")].value}')"
pvc_build_log="$(task_log "${pvc_recovery}" build)"
pvc_taskruns="$(k get taskrun -l "tekton.dev/pipelineRun=${pvc_recovery}" --no-headers | wc -l | tr -d ' ')"

# Negative case: a volumeClaimTemplate gives the recovery run a new, empty volume.
export_source retry-stateful-vct-source
vct_warnings="$(plan_warnings retry-stateful-vct-source)"
if "${tmp_dir}/retryctl" create --pipeline "${tmp_dir}/pipeline.yaml" --run "${tmp_dir}/retry-stateful-vct-source.yaml" \
  --policy integration/stateful-policy.yaml --taskruns "${tmp_dir}/retry-stateful-vct-source-taskruns.yaml" </dev/null >/dev/null 2>&1; then
  vct_unconfirmed="generated"
else
  vct_unconfirmed="stopped-for-confirmation"
fi
vct_recovery="$(recover retry-stateful-vct-source)"
k wait --for=condition=Succeeded=False --timeout=5m "${vct_recovery}"
vct_recovery="${vct_recovery#*/}"
vct_build_status="$(task_field "${vct_recovery}" build '{.items[0].status.conditions[0].status}')"
vct_build_log="$(task_log "${vct_recovery}" build)"

# Definition check: after the Pipeline changes, recovering the old run is refused.
k patch pipeline/retry-stateful --type=json \
  -p '[{"op":"replace","path":"/spec/tasks/3/taskSpec/steps/0/script","value":"echo \"deployed v2 $(params.image)\"\n"}]' >/dev/null
k get pipeline/retry-stateful -o yaml > "${tmp_dir}/pipeline.yaml"
"${tmp_dir}/retryctl" plan --pipeline "${tmp_dir}/pipeline.yaml" --run "${tmp_dir}/retry-stateful-pvc-source.yaml" \
  --policy integration/stateful-policy.yaml --taskruns "${tmp_dir}/retry-stateful-pvc-source-taskruns.yaml" > "${artifact_dir}/changed-definition-plan.json"
changed_decision="$(python3 -c 'import json,sys; p=json.load(open(sys.argv[1])); print(p["decision"] + ": " + p.get("refusalReason", ""))' "${artifact_dir}/changed-definition-plan.json")"

k get pipelinerun -o yaml > "${artifact_dir}/pipelineruns.yaml"
k get taskrun -o yaml > "${artifact_dir}/taskruns.yaml"

SOURCE_COMMIT="${source_commit}" RECOVERY_COMMIT="${recovery_commit}" PVC_WARNINGS="${pvc_warnings}" \
PVC_RECOVERY="${pvc_recovery}" PVC_TASKRUNS="${pvc_taskruns}" PVC_BUILD_LOG="${pvc_build_log}" \
VCT_WARNINGS="${vct_warnings}" VCT_UNCONFIRMED="${vct_unconfirmed}" VCT_RECOVERY="${vct_recovery}" \
VCT_BUILD_STATUS="${vct_build_status}" VCT_BUILD_LOG="${vct_build_log}" CHANGED_DECISION="${changed_decision}" \
python3 - "${artifact_dir}/summary.json" <<'EOF'
import json, os, sys
e = os.environ
summary = {
    "persistentClaim": {
        "recoveryRun": e["PVC_RECOVERY"],
        "recoverySucceeded": True,
        "recoveryTaskRuns": int(e["PVC_TASKRUNS"]),
        "warnings": e["PVC_WARNINGS"].split(",") if e["PVC_WARNINGS"] else [],
        "sourceCommitResult": e["SOURCE_COMMIT"],
        "substitutedCommitParam": e["RECOVERY_COMMIT"],
        "commitMatches": e["SOURCE_COMMIT"] == e["RECOVERY_COMMIT"] != "",
        "buildLog": e["PVC_BUILD_LOG"],
    },
    "volumeClaimTemplate": {
        "recoveryRun": e["VCT_RECOVERY"],
        "warnings": e["VCT_WARNINGS"].split(",") if e["VCT_WARNINGS"] else [],
        "withoutConfirmation": e["VCT_UNCONFIRMED"],
        "recoverySucceeded": False,
        "buildSucceeded": e["VCT_BUILD_STATUS"],
        "buildLog": e["VCT_BUILD_LOG"],
    },
    "changedDefinition": e["CHANGED_DECISION"],
}
json.dump(summary, open(sys.argv[1], "w"), indent=2)
print(json.dumps(summary, indent=2))
EOF
