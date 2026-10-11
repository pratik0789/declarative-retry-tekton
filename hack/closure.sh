#!/usr/bin/env bash
# Closure ablation: the same failure recovered with an empty policy and with the
# declared closure, judged against a full restart of the same revision.
# A recovery is correct when it succeeds and its build artifact equals the restart's.
# Chained check (retry-vendor): the failed empty-policy recovery run cannot itself be
# recovered, but its source can be recovered again under the declared closure.
set -euo pipefail

cluster="${CLUSTER:-retry-paper}"
context="kind-${cluster}"
namespace="retry-closure"
trials="${TRIALS:-5}"
artifact_dir="${ARTIFACT_DIR:-artifacts/closure}"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT
mkdir -p "${artifact_dir}"
k() { kubectl --context "${context}" -n "${namespace}" "$@"; }

go build -o "${tmp_dir}/retryctl" ./cmd/retryctl
kubectl --context "${context}" delete namespace "${namespace}" --ignore-not-found --wait=true
kubectl --context "${context}" create namespace "${namespace}"
k apply -f integration/closure.yaml
k run image-warmup --image=alpine:3.20 --restart=Never --command -- true
k wait --for=jsonpath='{.status.phase}'=Succeeded --timeout=2m pod/image-warmup

pvc() {
  cat <<EOF | k apply -f - >/dev/null
apiVersion: v1
kind: PersistentVolumeClaim
metadata: {name: $1}
spec:
  accessModes: [ReadWriteOnce]
  resources: {requests: {storage: 64Mi}}
EOF
}
start_run() {
  cat <<EOF | k create -f - >/dev/null
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata: {name: $1}
spec:
  pipelineRef: {name: $2}
  params: [{name: revision, value: $3}]
  workspaces: [{name: shared, persistentVolumeClaim: {claimName: $4}}]
EOF
}
wait_done() {
  local s
  for _ in $(seq 1 600); do
    s="$(k get "pipelinerun/$1" -o jsonpath='{.status.conditions[?(@.type=="Succeeded")].status}')"
    if [ "$s" = True ] || [ "$s" = False ]; then echo "$s"; return; fi
    sleep 1
  done
  echo Timeout
}
task_field() { k get taskrun -l "tekton.dev/pipelineRun=$1,tekton.dev/pipelineTask=$2" -o "jsonpath=$3" 2>/dev/null || true; }
build_log() {
  local pod
  pod="$(task_field "$1" build '{.items[0].status.podName}')"
  [ -n "${pod}" ] && k logs "pod/${pod}" --all-containers 2>/dev/null | grep -E '^STATE-' | tr '\n' ' ' || true
}
export_run() {
  k get "pipelinerun/$1" -o yaml > "${tmp_dir}/$1.yaml"
  k get taskrun -l "tekton.dev/pipelineRun=$1" -o yaml > "${tmp_dir}/$1-taskruns.yaml"
}
plan_json() { # pipeline source policy -> plan json path
  "${tmp_dir}/retryctl" plan --pipeline "${tmp_dir}/$1-pipeline.yaml" --run "${tmp_dir}/$2.yaml" \
    --policy "$3" --taskruns "${tmp_dir}/$2-taskruns.yaml" > "${artifact_dir}/$2-$(basename "$3" .yaml)-plan.json"
  echo "${artifact_dir}/$2-$(basename "$3" .yaml)-plan.json"
}
plan_field() { python3 -c 'import json,sys; p=json.load(open(sys.argv[1])); f=sys.argv[2]
if f=="warnings": print(";".join(sorted(w["code"] for w in p.get("warnings", []))))
elif f=="rerun": print(";".join(t["name"] for t in (p.get("tasks") or []) if t.get("action")=="rerun"))
else: print((p.get("decision","")+": "+p.get("refusalReason","")).rstrip(": "))' "$1" "$2"; }
recover() { # pipeline source policy -> recovery run name
  "${tmp_dir}/retryctl" create --pipeline "${tmp_dir}/$1-pipeline.yaml" --run "${tmp_dir}/$2.yaml" \
    --policy "$3" --taskruns "${tmp_dir}/$2-taskruns.yaml" --confirm-warnings \
    2>/dev/null > "${artifact_dir}/$2-$(basename "$3" .yaml)-recovery.yaml"
  local name
  name="$(k create -f "${artifact_dir}/$2-$(basename "$3" .yaml)-recovery.yaml" -o name)"
  echo "${name#*/}"
}

csv="${artifact_dir}/trials.csv"
echo 'pipeline,trial,arm,decision,rerunTasks,warnings,recoveryStatus,image,restartImage,correct,buildLog' > "${csv}"
chained="${artifact_dir}/chained.csv"
echo 'trial,failedRecoveryAsSource,reRecoveryStatus,image,restartImage,correct' > "${chained}"

for p in codegen vendor; do
  pipeline="retry-${p}"
  k get "pipeline/${pipeline}" -o yaml > "${tmp_dir}/${pipeline}-pipeline.yaml"
  for t in $(seq 1 "${trials}"); do
    rev="rev-${p}-${t}"
    for arm in empty closure restart; do pvc "${p}-${arm}-${t}"; done
    start_run "${p}-source-empty-${t}" "${pipeline}" "${rev}" "${p}-empty-${t}"
    start_run "${p}-source-closure-${t}" "${pipeline}" "${rev}" "${p}-closure-${t}"
    start_run "${p}-restart-${t}" "${pipeline}" "${rev}" "${p}-restart-${t}"
    [ "$(wait_done "${p}-source-empty-${t}")" = False ]
    [ "$(wait_done "${p}-source-closure-${t}")" = False ]
    [ "$(wait_done "${p}-restart-${t}")" = True ]
    restart_image="$(task_field "${p}-restart-${t}" build '{.items[0].status.results[?(@.name=="image")].value}')"

    for arm in empty closure; do
      src="${p}-source-${arm}-${t}"
      if [ "${arm}" = empty ]; then policy="integration/closure-${p}-empty-policy.yaml"; else policy="integration/closure-${p}-policy.yaml"; fi
      export_run "${src}"
      plan="$(plan_json "${pipeline}" "${src}" "${policy}")"
      run="$(recover "${pipeline}" "${src}" "${policy}")"
      status="$(wait_done "${run}")"
      image="$(task_field "${run}" build '{.items[0].status.results[?(@.name=="image")].value}')"
      correct=false
      if [ "${status}" = True ] && [ "${image}" = "${restart_image}" ]; then correct=true; fi
      echo "${pipeline},${t},${arm},$(plan_field "${plan}" decision),$(plan_field "${plan}" rerun),$(plan_field "${plan}" warnings),${status},${image},${restart_image},${correct},$(build_log "${run}")" >> "${csv}"

      if [ "${p}" = vendor ] && [ "${arm}" = empty ] && [ "${status}" = False ]; then
        export_run "${run}"
        chained_plan="$(plan_json "${pipeline}" "${run}" "integration/closure-${p}-policy.yaml")"
        export_run "${src}"
        rerun="$(recover "${pipeline}" "${src}" "integration/closure-${p}-policy.yaml")"
        rstatus="$(wait_done "${rerun}")"
        rimage="$(task_field "${rerun}" build '{.items[0].status.results[?(@.name=="image")].value}')"
        rcorrect=false
        if [ "${rstatus}" = True ] && [ "${rimage}" = "${restart_image}" ]; then rcorrect=true; fi
        echo "${t},$(plan_field "${chained_plan}" decision),${rstatus},${rimage},${restart_image},${rcorrect}" >> "${chained}"
      fi
    done
  done
done

k get pipelinerun -o yaml > "${artifact_dir}/pipelineruns.yaml"
k get taskrun -o yaml > "${artifact_dir}/taskruns.yaml"
python3 - "${csv}" "${chained}" "${artifact_dir}/summary.json" <<'EOF'
import csv, json, sys
rows = list(csv.DictReader(open(sys.argv[1])))
chained = list(csv.DictReader(open(sys.argv[2])))
summary = {}
for r in rows:
    s = summary.setdefault(r["pipeline"], {}).setdefault(r["arm"], {"trials": 0, "succeeded": 0, "correct": 0, "silentlyWrong": 0, "failed": 0, "warnings": set(), "rerun": set()})
    s["trials"] += 1
    s["succeeded"] += r["recoveryStatus"] == "True"
    s["correct"] += r["correct"] == "true"
    s["silentlyWrong"] += r["recoveryStatus"] == "True" and r["correct"] != "true"
    s["failed"] += r["recoveryStatus"] != "True"
    s["warnings"].add(r["warnings"]); s["rerun"].add(r["rerunTasks"])
for p in summary.values():
    for s in p.values():
        s["warnings"] = sorted(s["warnings"]); s["rerun"] = sorted(s["rerun"])
summary["chainedVendor"] = {
    "trials": len(chained),
    "failedRecoveryAsSourceDecisions": sorted({c["failedRecoveryAsSource"] for c in chained}),
    "reRecoveryCorrect": sum(c["correct"] == "true" for c in chained),
}
json.dump(summary, open(sys.argv[3], "w"), indent=2)
print(json.dumps(summary, indent=2))
EOF
