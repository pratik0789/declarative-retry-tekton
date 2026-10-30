# Tekton retry experiment: independent source-only adjudication

## Status and evidence boundary

This is an **automated independent source assessment, not a practitioner authoring study**. It was derived before access to experiment policies, planner outputs, outcomes, ledgers, manuscripts, or prior reviews. It uses only the external-service Python source and pipeline description supplied in the task. No repository or manuscript was modified. CLI documentation and risk annotations were consulted solely to publish this assessment.

Pipeline: `warmup -> acquire -> publish -> sign -> scan -> deploy`. Warmup prints a constant. Other tasks POST `{}` and fail on non-2xx. There are no Task results or Workspaces. Each source run has exactly one injected failure, at publish, sign, or scan; completed predecessors remain completed and successors have not run.

These are analytical predictions, not observations from a live Tekton run. They assume retry execution preserves serial task order, reuses completed tasks unless selected, runs the failed task and its previously unexecuted successors, and starts each strategy from the same captured external source state. The one-shot `injected=True` flag must remain true on restore. Epoch stays fixed throughout each retry. No administrative reset or hidden external writes occur during retry.

## Source-state reconstruction

The injection branch precedes the task implementation. The injected 503 sets `injected=True`, appends an event, and makes no credential/artifact/signature/scan/deploy write.

| Source failure | credential | artifact | signature | scanned | deployed | Completed source tasks |
|---|---|---|---|---|---|---|
| publish | 1 | null | null | null | null | warmup, acquire |
| sign | 1 | 1 | null | null | null | warmup, acquire, publish |
| scan | 1 | 1 | 1 | null | null | warmup, acquire, publish, sign |

Stable restores retain epoch=1. Rotated restores set epoch=2 while retaining all produced fields above at 1. Thus the two worlds have the same successful task history but different validity of external state.

## Minimum earlier-task reselections

These sets exclude the failed task and not-yet-run successors, which already need execution.

| Failure | Stable epoch 1 | Rotated epoch 2 | Necessity argument in rotated world |
|---|---|---|---|
| publish | empty set | {acquire} | Only acquire writes credential; publish requires credential=epoch. |
| sign | empty set | {acquire, publish} | Sign requires both current credential and current artifact. Only acquire and publish write them; publish itself needs current credential. |
| scan | empty set | {acquire, publish, sign} | Scan requires current signature. Only sign writes it, and sign requires current credential and artifact. |

Each rotated set is necessary and sufficient under the stated execution assumptions. For stable sources the prerequisite values already equal epoch, so adding any earlier task cannot reduce the number needed for success. Warmup cannot repair any prerequisite and is never necessary.

## Single conservative retryWith relation

Define `retryWith(t)` as earlier tasks to add when task t is selected for retry. A single explicit conservative relation is:

```text
publish: [acquire]
sign:    [acquire, publish]
scan:    [acquire, publish, sign]
acquire, warmup, deploy: []
```

This is the smallest per-failure explicit set assignment that supports both worlds without observing the epoch, when entries must directly name all required additions. It supports the specified publish/sign/scan failures; it does not claim completeness for all possible failures or external worlds. In stable worlds it intentionally reruns valid earlier work.

If the consumer recursively closes retryWith additions, the smaller edge presentation `publish -> acquire; sign -> publish; scan -> sign` induces exactly the same selections for these failures. The source alone does not specify recursive relation semantics; the explicit presentation avoids depending on them. These arrows express retry co-selection, not reversed execution order. Selected tasks must execute in original pipeline order.

## Expected executed-task counts

Count an attempted task even when its HTTP request fails. Counts below are for a **single retry attempt only**, include warmup when executed, and exclude completed tasks merely reused. Full restart reruns all six tasks against the restored external world without resetting epoch or reinjecting failure.

“Empty” has no earlier additions. “Declared” uses the conservative relation above. “Informed manual” knows the current external world and selects the minimum set above; it is an ideal comparator, not an observed human performance result. “Full” reruns all six tasks.

| Failure | World | Empty: actual executions/outcome | Declared: actual executions/outcome | Informed manual: actual executions/outcome | Full: actual executions/outcome |
|---|---|---|---|---|---|
| publish | stable | 4 / success | 5 / success | 4 / success | 6 / success |
| sign | stable | 3 / success | 5 / success | 3 / success | 6 / success |
| scan | stable | 2 / success | 5 / success | 2 / success | 6 / success |
| publish | rotated | 1 / failure, publish 409 | 5 / success | 5 / success | 6 / success |
| sign | rotated | 1 / failure, sign 409 | 5 / success | 5 / success | 6 / success |
| scan | rotated | 1 / failure, scan 409 | 5 / success | 5 / success | 6 / success |

In stable empty/manual retries, the executed suffixes are respectively publish/sign/scan/deploy (4), sign/scan/deploy (3), and scan/deploy (2). In every declared retry and rotated manual retry, the executed sequence is acquire/publish/sign/scan/deploy (5). Full restart additionally executes warmup (6).

Empty retry has candidate suffix sizes 4, 3, 2 in either world, but in rotated worlds only the first task executes because it fails immediately. Counting its unexecuted successors as executions would be an error.

Across one attempt per six failure/world cells: empty executes 12 tasks and succeeds in 3/6; declared executes 30 and succeeds in 6/6; ideal informed manual executes 24 and succeeds in 6/6; full restart executes 36 and succeeds in 6/6. These six deliberately constructed cells are not a population sample or statistical estimate. Unequal success makes the empty execution total unsuitable as a like-for-like successful-recovery cost comparison.

Declared saves one task relative to full in each cell (1/6 of full restart's executed tasks); manual saves 2, 3, 4 tasks in the respective stable cells and 1 task in each rotated cell. Such task counts do not establish wall-clock latency, resource cost, practitioner effort, or workload-general savings.

If counting original source execution plus retry, original attempts are 3 tasks for publish failure, 4 for sign, 5 for scan, including warmup and the injected failure. Add respectively 3, 4, 5 to each row above. Across all six cells the original executions total 24, producing combined totals 36/54/48/60 for empty/declared/manual/full. The retry-only and combined denominators must not be mixed.

## External final-state predicate

For a restored world's fixed epoch e, successful semantic recovery requires:

```text
state.epoch == e
AND state.credential == e
AND state.artifact == e
AND state.signature == e
AND state.scanned == e
AND state.deployed == e
```

Here e is independently fixed by the restore (1 or 2), rather than inferred from the planner. `deployed==e` alone follows the source's immediate deploy check on scanned, but the full predicate explicitly checks every represented service field and protects against accidental fixture contamination. The predicate is evaluated by reading the external service state, not by accepting retryctl predictions or PipelineRun success as ground truth.

Additional execution-audit checks are distinct from that semantic predicate: restored `injected=True`; no epoch changes; retry event suffix contains precisely the actually executed external-task sequence; expected HTTP statuses; and no hidden reset/restore during execution. Warmup generates no service event, so service-event counts are one smaller than task executions for a full restart (5 events, 6 tasks). For other successful strategies all executed tasks call the service. Restored source events must be excluded when counting retry events.

Oracle independence is limited: this service is separate from retryctl and has no Tekton API access, so its observed fields can check planner-independent effects. It is still an authored deterministic synthetic oracle whose behaviors define the experiment. This assessment does not establish realism, practical policy-authoring accuracy, independent human agreement, or correctness of any actual harness/controller.

## Ancestor and parallel validity

Every explicit added task is a strict ancestor of its target: acquire precedes publish; acquire/publish precede sign; acquire/publish/sign precede scan. No self-selection, successor selection, or parallel pair appears. Thus a strict-ancestor eligibility rule accepts the relation, and a rule permitting only ancestors or parallel tasks also accepts it via the ancestor case.

The supplied pipeline is a total serial chain, so there are **no parallel task pairs**. It cannot validate acceptance or rejection of parallel retryWith targets or execution under concurrency. No Task results or Workspaces means the prerequisite edges here are external-state requirements, not result-reference or workspace wiring dependencies. The pipeline graph locates valid ancestors but cannot reveal whether epoch-bound external effects remain valid. Ancestor admissibility alone does not prove a policy sufficient: the empty relation is structurally admissible and fails all three rotated worlds.

These conclusions are conditioned on fixed epoch and serial execution. Rotation during the retry could invalidate a just-written credential or artifact and is outside the described two-world experiment. The source overwrites fields and supplies no compensation, transactions, credential secrecy, real signing, scanning quality, or real deployment behavior; no inference about those mechanisms follows.

## Locked adjudication

The world-specific minimum earlier sets are empty/empty/empty in stable epoch 1 and {acquire}/{acquire,publish}/{acquire,publish,sign} in rotated epoch 2. A single conservative declaration runs five tasks in every cell; full restart runs six. Empty policy succeeds only in the three stable cells. These predictions were fixed before Stage B and before exposure to actual experimental artifacts.
