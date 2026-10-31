#!/usr/bin/env bash
set -euo pipefail

cluster="${CLUSTER:-retry-paper}"
context="kind-${cluster}"
namespace="${NAMESPACE:-retry-benchmark}"
sizes="${SIZES:-5 10}"
positions="${POSITIONS:-early middle late}"
repetitions="${REPETITIONS:-3}"
repetition_start="${REPETITION_START:-1}"
task_sleep="${TASK_SLEEP:-0.2}"
order_seed="${ORDER_SEED:-20261030}"
output_dir="${OUTPUT_DIR:-artifacts/benchmark}"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT
mkdir -p "${output_dir}"
results="${output_dir}/trials.csv"

go build -o "${tmp_dir}/retryctl" ./cmd/retryctl
if kubectl --context "${context}" get namespace "${namespace}" >/dev/null 2>&1; then
  echo "Namespace ${namespace} already exists; use a fresh NAMESPACE to preserve earlier evidence." >&2
  exit 2
fi
kubectl --context "${context}" create namespace "${namespace}"
sha256sum "${tmp_dir}/retryctl" > "${output_dir}/retryctl-sha256.txt"
go version > "${output_dir}/go-version.txt"
git rev-parse HEAD > "${output_dir}/parent-commit.txt"
kubectl --context "${context}" run image-warmup -n "${namespace}" --image=alpine:3.20 --restart=Never --command -- true
kubectl --context "${context}" wait -n "${namespace}" --for=jsonpath='{.status.phase}'=Succeeded --timeout=2m pod/image-warmup

/usr/bin/python3 - "${sizes}" "${positions}" "${repetition_start}" "${repetitions}" "${order_seed}" "${output_dir}" <<'PYORDER'
import json, random, sys
from pathlib import Path
sizes, positions, start, count, seed, dest = sys.argv[1:]
rng = random.Random(int(seed))
orders = {}
for size in sizes.split():
    for position in positions.split():
        sequence = ['baseline-first'] * (int(count)//2) + ['recovery-first'] * (int(count)//2)
        if int(count)%2: sequence.append(rng.choice(['baseline-first','recovery-first']))
        rng.shuffle(sequence)
        for repetition, order in zip(range(int(start), int(start)+int(count)), sequence):
            orders[f's{size}-{position[0]}-r{repetition}'] = order
root = Path(dest)
(root/'arm-order.json').write_text(json.dumps({'seed':int(seed),'method':'Within each size/position, balanced arm orders shuffled by Python random.Random before execution','orders':orders},indent=2)+'\n')
(root/'protocol.md').write_text('Current implementation timing replication. Synthetic serial tasks; fixed 0.2-second work. Ten paired repetitions for each of two sizes and three failure positions. Arm orders balanced and shuffled before any timing outcomes. No concurrent native experiment. Wall durations include create/wait commands; controller timestamps retained separately. Descriptive local overhead result, not operator effort or realistic long-task speedup. Failed/incomplete pairs retained.\n')
PYORDER

echo 'size,position,failureTask,repetition,baselineMs,recoveryMs,baselineTasks,recoveryTasks,tasksAvoided,taskAvoidancePercent,timeReductionPercent' > "${results}"

task_name() { printf 't%02d' "$1"; }

for size in ${sizes}; do
  for position in ${positions}; do
    case "${position}" in
      early) failure_index=2 ;;
      middle) failure_index=$(((size + 1) / 2)) ;;
      late) failure_index="${size}" ;;
      *) echo "unknown position: ${position}" >&2; exit 2 ;;
    esac
    failure_task="$(task_name "${failure_index}")"
    repetition_end=$((repetition_start + repetitions - 1))
    for repetition in $(seq "${repetition_start}" "${repetition_end}"); do
      case_id="s${size}-${position:0:1}-r${repetition}"
      pipeline_name="bench-${case_id}"
      source_name="${pipeline_name}-source"
      pipeline_file="${tmp_dir}/${case_id}-pipeline.yaml"
      policy_file="${tmp_dir}/${case_id}-policy.yaml"
      run_file="${tmp_dir}/${case_id}-run.yaml"
      taskruns_file="${tmp_dir}/${case_id}-taskruns.yaml"
      recovery_file="${tmp_dir}/${case_id}-recovery.yaml"

      {
        echo 'apiVersion: tekton.dev/v1'
        echo 'kind: Pipeline'
        echo 'metadata:'
        echo "  name: ${pipeline_name}"
        echo "  namespace: ${namespace}"
        echo 'spec:'
        echo '  tasks:'
        for index in $(seq 1 "${size}"); do
          name="$(task_name "${index}")"
          echo "    - name: ${name}"
          if [ "${index}" -gt 1 ]; then echo "      runAfter: [$(task_name "$((index - 1))")]"; fi
          echo '      taskSpec:'
          echo '        steps:'
          echo '          - name: work'
          echo '            image: alpine:3.20'
          echo '            script: |'
          echo "              sleep ${task_sleep}"
          if [ "${index}" -eq "${failure_index}" ]; then
            echo "              case \"\$(context.taskRun.name)\" in \"${source_name}\"-*) exit 1;; esac"
          fi
          echo "              echo ${name}"
        done
        echo '---'
        echo 'apiVersion: tekton.dev/v1'
        echo 'kind: PipelineRun'
        echo 'metadata:'
        echo "  name: ${source_name}"
        echo "  namespace: ${namespace}"
        echo 'spec:'
        echo '  pipelineRef:'
        echo "    name: ${pipeline_name}"
      } > "${pipeline_file}"

      {
        echo 'apiVersion: retry.tekton.dev/v1alpha1'
        echo 'kind: PipelineRetryPolicy'
        echo "metadata: {name: policy-${case_id}}"
        echo 'spec:'
        echo "  pipelineRef: ${pipeline_name}"
        echo '  tasks: {}'
      } > "${policy_file}"

      kubectl --context "${context}" apply -f "${pipeline_file}"
      kubectl --context "${context}" wait -n "${namespace}" --for=condition=Succeeded=False --timeout=10m "pipelinerun/${source_name}"
      kubectl --context "${context}" get pipeline "${pipeline_name}" -n "${namespace}" -o yaml > "${tmp_dir}/${case_id}-live-pipeline.yaml"
      kubectl --context "${context}" get pipelinerun "${source_name}" -n "${namespace}" -o yaml > "${run_file}"
      kubectl --context "${context}" get taskrun -n "${namespace}" -l "tekton.dev/pipelineRun=${source_name}" -o yaml > "${taskruns_file}"

      baseline_name="${pipeline_name}-full"
      "${tmp_dir}/retryctl" create --pipeline "${tmp_dir}/${case_id}-live-pipeline.yaml" --run "${run_file}" --policy "${policy_file}" --taskruns "${taskruns_file}" > "${recovery_file}"
      arm_order="$(/usr/bin/python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["orders"][sys.argv[2]])' "${output_dir}/arm-order.json" "${case_id}")"
      if [ "${arm_order}" = baseline-first ]; then arms="baseline recovery"; else arms="recovery baseline"; fi
      for arm in ${arms}; do
        arm_start="$(date +%s%3N)"
        if [ "${arm}" = baseline ]; then
          printf 'apiVersion: tekton.dev/v1\nkind: PipelineRun\nmetadata:\n  name: %s\nspec:\n  pipelineRef:\n    name: %s\n' "${baseline_name}" "${pipeline_name}" | kubectl --context "${context}" create -n "${namespace}" -f -
          arm_resource="pipelinerun/${baseline_name}"
        else
          recovery_resource="$(kubectl --context "${context}" create -n "${namespace}" -f "${recovery_file}" -o name)"
          recovery_name="${recovery_resource#*/}"
          arm_resource="${recovery_resource}"
        fi
        kubectl --context "${context}" wait -n "${namespace}" --for=condition=Succeeded=True --timeout=10m "${arm_resource}"
        arm_end="$(date +%s%3N)"
        if [ "${arm}" = baseline ]; then baseline_ms=$((arm_end - arm_start)); else recovery_ms=$((arm_end - arm_start)); fi
      done

      case_dir="${output_dir}/${case_id}"
      mkdir -p "${case_dir}"
      cp "${tmp_dir}/${case_id}-live-pipeline.yaml" "${case_dir}/pipeline.yaml"
      cp "${policy_file}" "${case_dir}/policy.yaml"
      cp "${run_file}" "${case_dir}/source-run.yaml"
      cp "${taskruns_file}" "${case_dir}/source-taskruns.yaml"
      cp "${recovery_file}" "${case_dir}/recovery.yaml"
      "${tmp_dir}/retryctl" plan --pipeline "${case_dir}/pipeline.yaml" --run "${case_dir}/source-run.yaml" \
        --policy "${case_dir}/policy.yaml" --taskruns "${case_dir}/source-taskruns.yaml" > "${case_dir}/plan.json"
      for run_name in "${baseline_name}" "${recovery_name}"; do
        kind=baseline
        if [ "${run_name}" = "${recovery_name}" ]; then kind=recovery; fi
        kubectl --context "${context}" get pipelinerun "${run_name}" -n "${namespace}" -o json > "${case_dir}/${kind}-run.json"
        kubectl --context "${context}" get taskrun -n "${namespace}" -l "tekton.dev/pipelineRun=${run_name}" -o json > "${case_dir}/${kind}-taskruns.json"
      done

      baseline_tasks="$(kubectl --context "${context}" get taskrun -n "${namespace}" -l "tekton.dev/pipelineRun=${baseline_name}" --no-headers | wc -l | tr -d ' ')"
      recovery_tasks="$(kubectl --context "${context}" get taskrun -n "${namespace}" -l "tekton.dev/pipelineRun=${recovery_name}" --no-headers | wc -l | tr -d ' ')"
      avoided=$((baseline_tasks - recovery_tasks))
      task_percent="$(awk -v a="${avoided}" -v b="${baseline_tasks}" 'BEGIN { printf "%.2f", 100*a/b }')"
      time_percent="$(awk -v r="${recovery_ms}" -v b="${baseline_ms}" 'BEGIN { printf "%.2f", 100*(b-r)/b }')"
      echo "${size},${position},${failure_task},${repetition},${baseline_ms},${recovery_ms},${baseline_tasks},${recovery_tasks},${avoided},${task_percent},${time_percent}" | tee -a "${results}"
    done
  done
done

awk -f benchmark/summarize.awk "${results}" | tee "${output_dir}/summary.csv"
kubectl --context "${context}" version -o yaml > "${output_dir}/kubernetes-version.yaml"
kubectl --context "${context}" get configmap pipelines-info -n tekton-pipelines -o yaml > "${output_dir}/tekton-version.yaml"
kubectl --context "${context}" get pipelinerun -n "${namespace}" -o json > "${output_dir}/all-pipelineruns.json"
kubectl --context "${context}" get taskrun -n "${namespace}" -o json > "${output_dir}/all-taskruns.json"
