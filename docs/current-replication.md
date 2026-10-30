# Current implementation replication

This revision retains fresh native evidence in `results/final-version-20261030/`.
The Go implementation is unchanged from the effective-policy identity revision.
Each experiment records binary/source hashes; the `parent-commit.txt` records the
checkout parent before these evidence/harness files were committed. It must not
be treated as an exact commit containing the new raw evidence. Compare the
retained file hashes with the final source instead.

Experiments run sequentially on the same local kind/Tekton cluster, in fresh
namespaces. Historical captures remain in their original directories. The
scripts refuse an existing namespace to avoid deleting old observations.

## Workspace replication

Seven repetitions of the original authored Workspace scenario retain native
PipelineRuns, TaskRuns, generated manifests, plans and state-verification logs.
Persistent-claim recovery succeeds; recovery with a fresh volumeClaimTemplate
loses state; changed-specification admission refuses. These are seven repetitions
of one authored application, not seven independent applications or an operator
study. `stateful-audit.json` checks raw native statuses and retained verification
logs. They do not prove general semantic safety.

```bash
NAMESPACE=unique-stateful-name ARTIFACT_DIR=output/stateful bash hack/stateful.sh
```

## Timing replication

The predeclared timing protocol uses two serial Pipeline sizes (5 and 10), three
failed-task positions, and ten restart/recovery pairs per cell. Each cell has five
restart-first and five recovery-first pairs, shuffled before execution with seed
20261030. The complete assignment is retained before any timing result. All
Tasks sleep 0.2 seconds, so startup and scheduling dominate. There is no concurrent
native experiment during these timings.

Wall durations include create/wait commands. Controller start/completion
measurements remain separately derivable from native records. These local short
Tasks do not estimate production latency or the value of authoring a policy.
Every empty policy retains the same failed-task-only choice an informed operator
could select.

```bash
NAMESPACE=unique-benchmark-name REPETITIONS=10 OUTPUT_DIR=output/benchmark bash hack/benchmark.sh
/usr/bin/python3 hack/audit_current_timing.py --directory output/benchmark
```

The audit checks actual successful terminal conditions, child TaskRun identities
and counts, nonoverlapping arm order, complete paired coverage, and retained
policy-content digests. It computes paired percentage reductions and Student-t
intervals across the ten repetitions, with individual negative pairs visible.

## Composition replications

Current-version ordering, finally-task and matched Workspace-closure replications
retain their own raw inputs, outputs and build identities. They confirm bounded
execution behavior of the current admission/generator implementation. The
closure predicates and policies remain project-authored fixtures; repeating them
does not measure author discovery, maintenance effort or operator error rates.

## Optional structural advisory and archived OCI elapsed times

`hack/check_reuse_convexity.py` is a separate optional Python companion, not a Go
planner change. It checks inherited work between executing ancestors and
descendants, including finally-phase order. Absence of its advisory does not
establish semantic reuse safety.

`public-oci-timing.json` derives controller elapsed seconds from the sixteen
already archived OCI executions. It adds no native run. The observations show
selected republish/rebuild tradeoffs and substantial variation; one observation
per cell cannot establish a general performance advantage over manual selection.

The exact used harness files are retained under `harnesses/`. After the timing
run, the public benchmark default was aligned to ten repetitions and its
protocol description made parameter-aware. Those description/default changes
do not alter the recorded run, which explicitly used ten repetitions and
0.2-second work. The retained used script remains available by hash.
