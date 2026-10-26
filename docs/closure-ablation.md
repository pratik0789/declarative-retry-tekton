# Matched stateful recovery comparisons

This experiment isolates `retryWith` from partial recovery in general. It runs
two author-created counterexamples on Tekton v1.6.0 / Kubernetes v1.31.4 / kind
v0.27.0, with five matched triplets per mechanism. These are feasibility
counterexamples, not a sample of production failures or policies from
independent authors.

## Experimental contract

One failed source PipelineRun and its TaskRuns are exported per triplet. Its
actual Workspace bytes are copied into three isolated persistent claims. The
harness checks equality before every arm. The empty-policy and closure arms
use the same source objects and generated recovery logic; the full-restart arm
uses the referenced Pipeline. Generated recovery manifests are retained before
the name/claimName override used to isolate effects. Arm order is shuffled with
seed 20261011. No timing benefit is claimed.

### Partial write

`observe -> prepare -> mutate -> verify -> publish`

The source `mutate` writes dirty input and fails. Retrying it alone sees that
dirty input and fails before another write. `mutate: retryWith: [prepare]`
reinitializes the input and lets recovery complete. The task's `DIRTY-INPUT`
check blocks the empty-policy arm; the planner itself only issues an advisory
Workspace warning.

### Stale build

`observe -> prepare -> build -> gate -> publish`

The source fails at `gate` after a successful build. The operator selects
`prepare`, whose input includes its TaskRun name and changes on each execution. An empty policy inherits the old
build; `publish` detects an input/artifact digest mismatch and fails before
writing a receipt. The failed `gate` declares
`gate: retryWith: [prepare, build]`, so recovery also regenerates the artifact.
Both named tasks precede `gate`. The operator's `--tasks prepare` selection is
retained in both partial-recovery arms; selecting successful earlier work is
explicit, and an operator must include any completed later work it invalidates.
`prepare` cannot declare the later `build` task in `retryWith`.
The planner does not itself certify semantic consistency.

## Output oracle

Native Tekton Succeeded conditions and TaskRuns report execution. A separate
Python check recomputes SHA-256 from captured Workspace bytes, checks the
artifact's stored input digest, checks artifact contents, and checks the
published receipt. This check uses actual bytes rather than expected planner
dispositions. Policies, tasks, fault mechanisms, and assertions nevertheless
share authorship; this is not independent policy adjudication.

Correct manual selection under an empty policy reproduces each declared
closure's task dispositions. The distinction is storing and applying recovery
knowledge, not adding an execution scope unavailable to a knowledgeable
operator.

## Reproduction

Use a new namespace; the harness never resets an existing namespace:

```bash
go build -trimpath -buildvcs=false -o /tmp/retryctl-closure ./cmd/retryctl
python3 hack/closure_ablation.py --context kind-retry-paper \
  --namespace retry-closure-new-unique-name --repetitions 5 \
  --retryctl /tmp/retryctl-closure \
  --output artifacts/closure-ablation
python3 hack/verify_closure_ablation.py \
  --results artifacts/closure-ablation --retryctl /tmp/retryctl-closure
```

If a transient cluster error interrupts a run, repeat the same harness command
with `--resume`. It verifies the saved cluster, namespace, and protocol, checks
existing resource specifications, preserves completed arms, and continues the
remaining work. Reads retry transient errors; creation requests are not retried
blindly. The October 25 run resumed after one etcd read timeout with 18 arms
already complete; its completed evidence was retained.

The current harness injects source-only failures with
`$(context.taskRun.name)` and the pattern `*-source-*`. TaskRun context is
resolved when the task executes, so the original expression remains in the
source PipelineRun's saved PipelineSpec. Stale-build input also uses the
prepare TaskRun name. This preserves the fault mechanisms and input refresh
while allowing the complete native saved/current specifications to match.
The captures are never rewritten to manufacture a match.

The verifier checks that every native specification matches, reruns the manual
selection comparisons with the evaluated binary, and rejects changed scripts
and changed context expressions in both planning and generation. Source,
harness, and binary hashes are retained in `build-metadata.json`. CI executes
one triplet per mechanism with the current implementation and retains its
evidence; the paper experiment uses five per mechanism.

The live harness and verifier use the Python standard library. The historical
numerical audit additionally requires PyYAML and SciPy. Retained evidence is in
`results/closure-current-20261029/` and
`results/current-version-audit-20261029/`. The current run enforces the direction
rule and the strict Pipeline-specification guard.
Earlier direction-only evidence in `results/closure-direction-20261025/`
contains resolved PipelineRun context expressions and is refused by the current
guard. It remains historical evidence, rather than the basis for the current
IV-G claims.
Historical evidence in `results/closure-ablation-20261011/` used an earlier
contract with `prepare: retryWith: [build]`; those artifacts are preserved
unchanged and do not validate the new restriction.

## Observed results (2026-10-29)

| Mechanism | Arm | Successful / trials | Consistent receipts / trials | TaskRuns per trial |
|---|---|---:|---:|---:|
| Partial write | Empty | 0/5 | 0/5 | 1, then failure |
| Partial write | Closure | 5/5 | 5/5 | 4 |
| Partial write | Full restart | 5/5 | 5/5 | 5 |
| Stale build | Empty | 0/5 | 0/5 | 3, including failed publish |
| Stale build | Closure | 5/5 | 5/5 | 4 |
| Stale build | Full restart | 5/5 | 5/5 | 5 |

Native objects and captured bytes verify all 30 arms. Correct manual selection
reproduces all ten declared plans. Each successful recovery with the closure avoids 1/5 TaskRuns
(20%) against restart; lower counts in failing arms are not successful savings.

## Limits

- Five repetitions per mechanism on one small Pipeline shape and one host.
- Synthetic local artifacts, not a registry, database, or deployment system.
- No policy-author study, failure-prevalence estimate, CPU/cost measurement,
  or production recovery-rate claim.
- Existing claims were forked for isolation; prior stateful results separately
  exercise mounting the original source claim.
- The new later-claim warning is tested on planning fixtures, not as a
  transaction or a live concurrent/chained recovery guarantee.
