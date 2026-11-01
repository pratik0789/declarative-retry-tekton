#!/bin/bash
# usage: identity.sh <namespace> <outdir>   (OLD_RETRYCTL, NEW_RETRYCTL)
set -euo pipefail
ns=$1; out=$2; ctx=kind-retry-paper; OLD=${OLD_RETRYCTL:?}; NEW=${NEW_RETRYCTL:?}
k() { kubectl --context $ctx -n $ns "$@"; }
mkdir -p $out
kubectl --context $ctx create namespace $ns
k apply -f integration/params.yaml
k get pvc retry-params-pvc -o json | python3 -c "import json,sys;d=json.load(sys.stdin);m=d['metadata'];[m.pop(x,None) for x in ('uid','resourceVersion','creationTimestamp','annotations','finalizers','managedFields')];m['name']='retry-params-pvc-2';d.pop('status',None);d['spec'].pop('volumeName',None);print(json.dumps(d))" | k create -f -
k get pipeline/retry-params -o yaml > $out/pipeline.yaml
spec=$(k get pipeline/retry-params -o json | python3 -c "import json,sys;print(json.dumps(json.load(sys.stdin)['spec']))")
cat <<YAML | k create -f -
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata: {name: resolver-source}
spec:
  pipelineRef:
    resolver: cluster
    params: [{name: kind, value: pipeline}, {name: name, value: retry-params}, {name: namespace, value: $ns}]
  params: [{name: release, value: "v1.4.2"}]
  workspaces: [{name: shared, persistentVolumeClaim: {claimName: retry-params-pvc}}]
YAML
k wait --for=condition=Succeeded=False --timeout=5m pipelinerun/resolver-source
python3 - "$spec" <<'PY' | k create -f -
import json,sys
print(json.dumps({"apiVersion":"tekton.dev/v1","kind":"PipelineRun","metadata":{"name":"embedded-source"},
 "spec":{"pipelineSpec":json.loads(sys.argv[1]),"params":[{"name":"release","value":"v1.4.2"}],
 "workspaces":[{"name":"shared","persistentVolumeClaim":{"claimName":"retry-params-pvc-2"}}]}}))
PY
k wait --for=condition=Succeeded=False --timeout=5m pipelinerun/embedded-source
summary="{}"
for src in resolver-source embedded-source; do
  k get pipelinerun/$src -o yaml > $out/$src-run.yaml
  k get taskrun -l tekton.dev/pipelineRun=$src -o yaml > $out/$src-taskruns.yaml
  args="--pipeline $out/pipeline.yaml --run $out/$src-run.yaml --taskruns $out/$src-taskruns.yaml --policy integration/params-policy.yaml"
  $OLD plan $args > $out/$src-plan-old.json
  $NEW plan $args > $out/$src-plan-new.json
  $NEW create $args --confirm-warnings 2>/dev/null > $out/$src-recovery.yaml
  name=$(k create -f $out/$src-recovery.yaml -o name)
  k wait --for=condition=Succeeded=True --timeout=5m $name
  pod=$(k get taskrun -l tekton.dev/pipelineRun=${name#*/},tekton.dev/pipelineTask=build -o jsonpath='{.items[0].status.podName}')
  k logs $pod --all-containers | grep STATE- > $out/$src-build.log || true
  k get taskrun -l tekton.dev/pipelineRun=${name#*/} -o yaml > $out/$src-recovery-taskruns.yaml
done
python3 - "$out" <<'PY'
import json,sys
o=sys.argv[1]; s={}
for src in ("resolver-source","embedded-source"):
    old=json.load(open(f"{o}/{src}-plan-old.json")); new=json.load(open(f"{o}/{src}-plan-new.json"))
    s[src]={"old":old.get("refusalReason") or old["decision"],"new":new["decision"],
            "rerun":sorted(t["name"] for t in new["tasks"] if t["action"]=="rerun"),
            "build":open(f"{o}/{src}-build.log").read().strip()}
json.dump(s,open(f"{o}/summary.json","w"),indent=1); print(json.dumps(s))
PY
