#!/bin/bash
# usage: run.sh <namespace> <outdir>
set -euo pipefail
ns=$1; out=$2; ctx=kind-retry-paper; OLD=${OLD_RETRYCTL:?}; NEW=${NEW_RETRYCTL:?}
k() { kubectl --context $ctx -n $ns "$@"; }
mkdir -p $out
kubectl --context $ctx create namespace $ns
k apply -f integration/params.yaml
cat <<YAML | k create -f -
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata: {name: retry-params-source}
spec:
  pipelineRef: {name: retry-params}
  params: [{name: release, value: "v1.4.2"}]
  workspaces: [{name: shared, persistentVolumeClaim: {claimName: retry-params-pvc}}]
YAML
k wait --for=condition=Succeeded=False --timeout=5m pipelinerun/retry-params-source
k get pipeline/retry-params -o yaml > $out/pipeline.yaml
k get pipelinerun/retry-params-source -o yaml > $out/source-run.yaml
k get taskrun -l tekton.dev/pipelineRun=retry-params-source -o yaml > $out/source-taskruns.yaml
args="--pipeline $out/pipeline.yaml --run $out/source-run.yaml --taskruns $out/source-taskruns.yaml --policy integration/params-policy.yaml"
$OLD plan $args > $out/plan-910cf54.json
$NEW plan $args > $out/plan-subst.json
$NEW create $args --confirm-warnings 2>/dev/null > $out/recovery.yaml
name=$(k create -f $out/recovery.yaml -o name)
k wait --for=condition=Succeeded=True --timeout=5m $name
k get $name -o yaml > $out/recovery-run.yaml
k get taskrun -l tekton.dev/pipelineRun=${name#*/} -o yaml > $out/recovery-taskruns.yaml
for t in build deploy; do pod=$(k get taskrun -l tekton.dev/pipelineRun=${name#*/},tekton.dev/pipelineTask=$t -o jsonpath='{.items[0].status.podName}'); k logs $pod --all-containers | grep -E "STATE-|DEPLOYED" > $out/recovery-$t.log || true; done
# changed definition: edit deploy script
k patch pipeline/retry-params --type=json -p '[{"op":"replace","path":"/spec/tasks/3/taskSpec/steps/0/script","value":"echo \"DEPLOYED v2 $(params.image)\"\n"}]' >/dev/null
k get pipeline/retry-params -o yaml > $out/pipeline-changed.yaml
$NEW plan --pipeline $out/pipeline-changed.yaml --run $out/source-run.yaml --taskruns $out/source-taskruns.yaml --policy integration/params-policy.yaml > $out/plan-changed-subst.json
python3 - "$out" <<'PY'
import json,sys
o=sys.argv[1]; L=lambda f: json.load(open(f"{o}/{f}"))
s={"old":L("plan-910cf54.json").get("refusalReason") or L("plan-910cf54.json")["decision"],
   "new":L("plan-subst.json")["decision"],
   "rerun":sorted(t["name"] for t in L("plan-subst.json")["tasks"] if t["action"]=="rerun"),
   "changed":L("plan-changed-subst.json").get("refusalReason"),
   "build_log":open(f"{o}/recovery-build.log").read().strip(),"deploy_log":open(f"{o}/recovery-deploy.log").read().strip()}
json.dump(s,open(f"{o}/summary.json","w"),indent=1); print(json.dumps(s))
PY
