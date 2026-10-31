# Tekton revision: independent source-only policy adjudication

## Status, provenance, and scope

This is an **automated independent source assessment, not a practitioner authoring study**. It is Stage A of the dedicated paper-review task, produced before receiving the finalized manuscript or experiment evidence. Its input is exclusively the external-service Python source and serial pipeline description in the user's initial task. No manuscript, prior review, generated policy, planner output, experiment result, or full live.py was consulted. No paper or repository files were modified. Administrative CLI documentation is outside the scientific input and is used only to publish this private artifact.

The source-only derivation below is a conditional prediction, not a report of observed Tekton executions. An independently written small state-transition model enumerated every subset of completed earlier tasks and reproduced all six unique minimum sets. That calculation checks the derivation; it is not an external experiment or human-authored policy evaluation.

## 1. Definitions and assumptions

Pipeline order: warmup -> acquire -> publish -> sign -> scan -> deploy. Warmup prints a constant and has no service effect. Each other task performs exactly one POST with '{}' and fails on non-2xx. There are no Task results or Workspaces. The injected publish/sign/scan failure is the first branch in do_POST: it marks injected=true, returns 503, and does not perform that task's ordinary state update. Previously successful tasks have completed; the suffix has not run.

For each captured source, stable means epoch=1. Rotated means epoch=2 with already-produced credential/artifact/signature still equal to 1. The consumed injection flag must remain true after restore; otherwise this is a different experiment with another injected failure. The epoch remains fixed during each retry, restore is complete before retry, and there are no concurrent writers.

"Earlier reselection" means choosing previously completed tasks to run again. The failed task and every not-yet-run suffix task are also executed, in original serial order; other successful earlier tasks are preserved/skipped. This defines the retry baseline semantics needed for the counts. They cannot be inferred from the HTTP service alone. A runtime that executes a different task selection is outside these predictions.

## 2. Captured states and minimum earlier reselections

| Injected failure | Produced fields at capture (epoch 1) | Stable minimum earlier set | Rotated minimum earlier set |
|---|---|---|---|
| publish | credential=1; artifact/signature/scanned/deployed=null | {} | {acquire} |
| sign | credential=artifact=1; signature/scanned/deployed=null | {} | {acquire, publish} |
| scan | credential=artifact=signature=1; scanned/deployed=null | {} | {acquire, publish, sign} |

These minima are unique under the stated retry semantics. In the stable world, every precondition for the failed task already holds and the one-time injection has been consumed. Thus retrying the failed task and suffix suffices.

For rotated publish, credential must be 2; acquire is its only producer. For rotated sign, both credential and artifact must be 2; acquire and publish are their only producers. For rotated scan, signature must be 2; sign is its only producer, and sign in turn requires credential=artifact=2. Hence acquire, publish, and sign must all rerun. No smaller earlier subset can succeed. Warmup never changes a precondition.

## 3. One conservative retryWith relation

Use the following semantic relation, with direction **failed task -> additional completed tasks to reselect**:

```
retryWith(publish) = {acquire}
retryWith(sign)    = {acquire, publish}
retryWith(scan)    = {acquire, publish, sign}
```

All other entries may be empty for the three failure cases under adjudication. This is a relation specification, not a claim about an unseen policy file's schema. It makes the complete necessary sets explicit and does not depend on recursive interpretation of retryWith entries. An immediate-predecessor chain could encode the same closure only if an implementation explicitly specifies and computes that recursive closure; that implementation has not been inspected.

The relation is conservative across both external worlds. Each entry equals the rotated minimum and overselects relative to the stable minimum by 1, 2, or 3 tasks. Its selected execution is acquire -> publish -> sign -> scan -> deploy in all six cases. It is the minimum world-independent successful selection for each failure under these assumptions; "conservative" does not mean arbitrary full restart.

The same captured Tekton task statuses/graph are compatible with different external epochs and therefore different minimum selections. Without external-state observations or supplied policy knowledge, these structural facts alone do not determine the minimum service-correct repair.

## 4. Expected retry executions

"Empty" uses no additional completed tasks. "Declared" below means exactly the conservative relation derived in Section 3, not any generated or observed experiment policy. "Informed manual" knows the external world and uses its minimum earlier set. "Full restart" runs all six tasks, including warmup, from the restored external state without another /reset or fresh injection. Count actual task executions, including a task whose POST fails. Do not count preserved/skipped tasks as executed. Administrative /restore and warmup are not /execute service events.

| Failure | World | Empty: actual tasks / outcome | Declared: tasks / outcome | Informed manual: tasks / outcome | Full restart: tasks / outcome |
|---|---|---|---|---|---|
| publish | stable | 4 / success | 5 / success | 4 / success | 6 / success |
| sign | stable | 3 / success | 5 / success | 3 / success | 6 / success |
| scan | stable | 2 / success | 5 / success | 2 / success | 6 / success |
| publish | rotated | 1 / failure (409 at publish) | 5 / success | 5 / success | 6 / success |
| sign | rotated | 1 / failure (409 at sign) | 5 / success | 5 / success | 6 / success |
| scan | rotated | 1 / failure (409 at scan) | 5 / success | 5 / success | 6 / success |

For rotated empty runs the initially selected suffix contains 4, 3, or 2 tasks respectively, but only its first task executes before failure. Comparing "planned selection size" to successful executed-task count would conflate different quantities. If an executor continues the suffix after a failed task, the actual counts change; the stated serial failure semantics presume stopping at that failure.

HTTP /execute event counts equal the task counts for empty, declared, and informed-manual retry. Full restart has six task executions and five /execute events because warmup makes no POST. These counts concern invocation count, not measured latency, monetary cost, or scalability.

The original captured source already executed 3 tasks for publish failure, 4 for sign failure, and 5 for scan failure, including warmup and the failed task. If counting source plus one retry, add those quantities to the corresponding cells:

| Failure | World | Empty cumulative tasks | Declared cumulative tasks | Informed manual cumulative tasks | Full restart cumulative tasks |
|---|---|---|---|---|---|
| publish | stable | 7 | 8 | 7 | 9 |
| sign | stable | 7 | 9 | 7 | 10 |
| scan | stable | 7 | 10 | 7 | 11 |
| publish | rotated | 4 (failed) | 8 | 8 | 9 |
| sign | rotated | 5 (failed) | 9 | 9 | 10 |
| scan | rotated | 6 (failed) | 10 | 10 | 11 |

Across the six cases, retry-only actual totals are empty=12 with only 3/6 successes, declared=30 with 6/6 successes, informed manual=24 with 6/6 successes, full restart=36 with 6/6 successes. These are deterministic predictions on a six-case constructed base. Empty's lower invocation count is not an equivalent successful-recovery baseline. Declared saves one task per successful case versus full restart (6/36 = 16.7% fewer task invocations in the six-case aggregate), entirely warmup; both make five service POSTs per case. Declared adds six task invocations versus the world-informed minimum aggregate (30 versus 24).

## 5. External final-state predicate and observable checks

Let E be the chosen world's expected epoch, fixed before executing the retry (E=1 for stable, E=2 for rotated). Read service state using GET after the retry terminates. The external recovery predicate is:

```
state.epoch == E
and state.credential == E
and state.artifact == E
and state.signature == E
and state.scanned == E
and state.deployed == E
```

Binding E externally prevents a harness from accidentally switching back to epoch 1 and satisfying a weaker all-fields-equal-to-current-epoch check. The predicate is independent of planner predictions, policy annotations, and Tekton completion status. It is service-state correctness within this deterministic integer-token model; it does not establish real credential freshness, cryptographic validity, artifact content integrity, or production safety.

Separately verify that injected stayed true, fail_at stayed the intended task, restored source fields matched the capture, and no extra reset/restore occurred during retry. Compare the **event suffix after the captured event prefix** with the expected ordered /execute invocations and HTTP codes; the restored events include the original source execution and its 503. Successful retries should produce only 200 responses; empty rotated retries should produce one 409 and leave scanned/deployed null. These are trace-integrity checks in addition to the final-state predicate. A final-state snapshot alone does not prove exact execution counts or which tasks produced it.

## 6. Ancestor and parallel validity

Every reselected task in the explicit relation is a strict transitive ancestor of its failed task in the given serial DAG. The relation has no self-edge, descendant edge, cycle, or nonexistent task. It is therefore ancestor-valid if the policy rule permits transitive ancestors. If a validator permits only immediate predecessors, the multi-task sign and scan entries require a different representation or closure rule; no such validator specification was supplied.

There are **no parallel/incomparable task pairs** in this pipeline: reachability orders every pair. A rule allowing only ancestors or parallel peers accepts these selections through the ancestor case, but this experiment provides no positive or negative test of parallel-peer handling. Parallel coverage cannot be inferred from the six cases. Structural admissibility also does not establish service-level sufficiency: a valid empty relation fails after rotation, and ancestor sets omitting required producers fail even though their individual edges are admissible.

No result or Workspace dependency communicates the epoch-bearing external state. The graph supplies execution order; the supplied service source supplies the hidden repair prerequisites. Warmup is an ancestor but semantically irrelevant, so selecting all ancestors would be sound here and unnecessarily execute one extra task.

## 7. Limits of the adjudication

The derivation assumes complete snapshot/restore, one consumed injected failure, fixed epoch during retry, sequential execution, no other writers, and the task-selection semantics defined above. It covers exactly the three failure locations and two external worlds. It predicts neither performance nor real-world prevalence, and it does not validate an implementation or measure practitioner policy authoring. Stage B will compare the finalized manuscript and immutable experiment evidence against these independently established expectations after the coordinator's ready message arrives.
