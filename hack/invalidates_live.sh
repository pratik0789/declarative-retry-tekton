#!/bin/bash
# usage: invalidates.sh <namespace> <outdir> <arm: empty|invalidates>   (NEW_RETRYCTL)
set -euo pipefail
ns=$1; out=$2; arm=$3; ctx=kind-retry-paper; NEW=${NEW_RETRYCTL:?}; IN=${INPUTS:?}
k() { kubectl --context $ctx -n $ns "$@"; }
mkdir -p $out
kubectl --context $ctx create namespace $ns
cat <<YAML | k create -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata: {name: data}
spec: {accessModes: [ReadWriteOnce], resources: {requests: {storage: 64Mi}}}
YAML
k create -f $IN/pipeline.json
cat <<YAML | k create -f -
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata: {name: sb-source}
spec:
  pipelineRef: {name: stale-build}
  workspaces: [{name: shared, persistentVolumeClaim: {claimName: data}}]
YAML
k wait --for=condition=Succeeded=False --timeout=5m pipelinerun/sb-source
k get pipeline/stale-build -o yaml > $out/pipeline.yaml
k get pipelinerun/sb-source -o yaml > $out/source-run.yaml
k get taskrun -l tekton.dev/pipelineRun=sb-source -o yaml > $out/source-taskruns.yaml
args="--pipeline $out/pipeline.yaml --run $out/source-run.yaml --taskruns $out/source-taskruns.yaml --policy $IN/policy-$arm.json --tasks prepare"
$NEW plan $args > $out/plan.json
$NEW create $args --confirm-warnings 2>/dev/null > $out/recovery.yaml
name=$(k create -f $out/recovery.yaml -o name)
for _ in $(seq 1 150); do st=$(k get $name -o jsonpath='{.status.conditions[?(@.type=="Succeeded")].status}'); [ "$st" = True ] || [ "$st" = False ] && break; sleep 2; done
k get $name -o yaml > $out/recovery-run.yaml
k get taskrun -l tekton.dev/pipelineRun=${name#*/} -o yaml > $out/recovery-taskruns.yaml
pod=$(k get taskrun -l tekton.dev/pipelineRun=${name#*/},tekton.dev/pipelineTask=publish -o jsonpath='{.items[0].status.podName}' 2>/dev/null || true)
[ -n "$pod" ] && k logs $pod --all-containers 2>/dev/null | grep -E "STALE|BAD|ORACLE" > $out/publish.log || true
python3 - "$out" <<'PY'
import json,sys,subprocess
o=sys.argv[1]; p=json.load(open(f"{o}/plan.json"))
import re
st=open(f"{o}/recovery-run.yaml").read()
ok='status: "True"' in st.split('conditions:')[1][:400] if 'conditions:' in st else False
s={"rerun":sorted(t["name"] for t in p["tasks"] if t["action"]=="rerun"),"warnings":[w["code"] for w in p.get("warnings") or []],
   "succeeded":ok,"publish":open(f"{o}/publish.log").read().strip() if __import__('os').path.exists(f"{o}/publish.log") else ""}
json.dump(s,open(f"{o}/summary.json","w"),indent=1); print(json.dumps(s))
PY
