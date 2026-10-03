#!/usr/bin/env bash
set -euo pipefail

cluster="${CLUSTER:-retry-paper}"
context="kind-${cluster}"
namespace="retry-eval"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT
artifact_dir="${ARTIFACT_DIR:-artifacts/integration}"
mkdir -p "${artifact_dir}"

kubectl --context "${context}" delete namespace "${namespace}" --ignore-not-found --wait=true
kubectl --context "${context}" create namespace "${namespace}"
kubectl --context "${context}" apply -f integration/source.yaml
kubectl --context "${context}" wait --for=condition=Succeeded=False --timeout=5m pipelinerun/retry-integration-source -n "${namespace}"
kubectl --context "${context}" get pipeline/retry-integration -n "${namespace}" -o yaml > "${tmp_dir}/pipeline.yaml"
kubectl --context "${context}" get pipelinerun/retry-integration-source -n "${namespace}" -o yaml > "${tmp_dir}/run.yaml"
kubectl --context "${context}" get taskrun -n "${namespace}" -l tekton.dev/pipelineRun=retry-integration-source -o yaml > "${tmp_dir}/taskruns.yaml"

baseline_start="$(date +%s%3N)"
cat <<'EOF' | kubectl --context "${context}" create -n "${namespace}" -f -
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  name: retry-integration-full
spec:
  pipelineRef:
    name: retry-integration
EOF
kubectl --context "${context}" wait --for=condition=Succeeded=True --timeout=5m pipelinerun/retry-integration-full -n "${namespace}"
baseline_end="$(date +%s%3N)"

go run ./cmd/retryctl create \
  --pipeline "${tmp_dir}/pipeline.yaml" \
  --run "${tmp_dir}/run.yaml" \
  --policy integration/policy.yaml \
  --taskruns "${tmp_dir}/taskruns.yaml" > "${tmp_dir}/recovery.yaml"

recovery_start="$(date +%s%3N)"
recovery_name="$(kubectl --context "${context}" create -n "${namespace}" -f "${tmp_dir}/recovery.yaml" -o name)"
kubectl --context "${context}" wait --for=condition=Succeeded=True --timeout=5m "${recovery_name}" -n "${namespace}"
recovery_end="$(date +%s%3N)"
recovery_short="${recovery_name#*/}"

kubectl --context "${context}" get pipelinerun -n "${namespace}" -o yaml > "${artifact_dir}/pipelineruns.yaml"
kubectl --context "${context}" get taskrun -n "${namespace}" -o yaml > "${artifact_dir}/taskruns.yaml"
kubectl --context "${context}" get events -n "${namespace}" --sort-by=.lastTimestamp > "${artifact_dir}/events.txt"
cp "${tmp_dir}/recovery.yaml" "${artifact_dir}/recovery.yaml"

baseline_tasks="$(kubectl --context "${context}" get taskrun -n "${namespace}" -l tekton.dev/pipelineRun=retry-integration-full --no-headers | wc -l | tr -d ' ')"
recovery_tasks="$(kubectl --context "${context}" get taskrun -n "${namespace}" -l "tekton.dev/pipelineRun=${recovery_short}" --no-headers | wc -l | tr -d ' ')"
cat > "${artifact_dir}/summary.json" <<EOF
{
  "sourceRun": "retry-integration-source",
  "baselineRun": "retry-integration-full",
  "recoveryRun": "${recovery_short}",
  "baselineSucceeded": true,
  "recoverySucceeded": true,
  "baselineTaskRuns": ${baseline_tasks},
  "recoveryTaskRuns": ${recovery_tasks},
  "tasksAvoided": $((baseline_tasks - recovery_tasks)),
  "baselineDurationMs": $((baseline_end - baseline_start)),
  "recoveryDurationMs": $((recovery_end - recovery_start))
}
EOF
cat "${artifact_dir}/summary.json"
