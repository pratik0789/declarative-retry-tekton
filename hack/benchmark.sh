#!/usr/bin/env bash
set -euo pipefail

cluster="${CLUSTER:-retry-paper}"
context="kind-${cluster}"
namespace="retry-benchmark"
sizes="${SIZES:-5 10}"
positions="${POSITIONS:-early middle late}"
repetitions="${REPETITIONS:-3}"
repetition_start="${REPETITION_START:-1}"
task_sleep="${TASK_SLEEP:-0.2}"
output_dir="${OUTPUT_DIR:-artifacts/benchmark}"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT
mkdir -p "${output_dir}"
results="${output_dir}/trials.csv"

go build -o "${tmp_dir}/retryctl" ./cmd/retryctl
kubectl --context "${context}" delete namespace "${namespace}" --ignore-not-found --wait=true
kubectl --context "${context}" create namespace "${namespace}"
kubectl --context "${context}" run image-warmup -n "${namespace}" --image=alpine:3.20 --restart=Never --command -- true
kubectl --context "${context}" wait -n "${namespace}" --for=jsonpath='{.status.phase}'=Succeeded --timeout=2m pod/image-warmup

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
            echo "              if [ \"\$(context.pipelineRun.name)\" = \"${source_name}\" ]; then exit 1; fi"
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
      baseline_start="$(date +%s%3N)"
      printf 'apiVersion: tekton.dev/v1\nkind: PipelineRun\nmetadata:\n  name: %s\nspec:\n  pipelineRef:\n    name: %s\n' "${baseline_name}" "${pipeline_name}" | kubectl --context "${context}" create -n "${namespace}" -f -
      kubectl --context "${context}" wait -n "${namespace}" --for=condition=Succeeded=True --timeout=10m "pipelinerun/${baseline_name}"
      baseline_end="$(date +%s%3N)"

      "${tmp_dir}/retryctl" create --pipeline "${tmp_dir}/${case_id}-live-pipeline.yaml" --run "${run_file}" --policy "${policy_file}" --taskruns "${taskruns_file}" > "${recovery_file}"
      recovery_start="$(date +%s%3N)"
      recovery_resource="$(kubectl --context "${context}" create -n "${namespace}" -f "${recovery_file}" -o name)"
      kubectl --context "${context}" wait -n "${namespace}" --for=condition=Succeeded=True --timeout=10m "${recovery_resource}"
      recovery_end="$(date +%s%3N)"
      recovery_name="${recovery_resource#*/}"

      baseline_tasks="$(kubectl --context "${context}" get taskrun -n "${namespace}" -l "tekton.dev/pipelineRun=${baseline_name}" --no-headers | wc -l | tr -d ' ')"
      recovery_tasks="$(kubectl --context "${context}" get taskrun -n "${namespace}" -l "tekton.dev/pipelineRun=${recovery_name}" --no-headers | wc -l | tr -d ' ')"
      baseline_ms=$((baseline_end - baseline_start))
      recovery_ms=$((recovery_end - recovery_start))
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
