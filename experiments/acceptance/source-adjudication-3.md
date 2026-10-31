# Tekton retry experiment: independent source-only adjudication

## Status and boundary

This is an **automated independent source assessment, not a practitioner authoring study**. It derives expectations before access to experimental policies, plans, outcomes, ledgers, or the manuscript. It is not an empirical report of Tekton executions, an evaluation of policy authoring by humans, or a conference acceptance verdict.

Inputs are exclusively the service source and pipeline description supplied in the task: serial `warmup -> acquire -> publish -> sign -> scan -> deploy`; warmup prints a constant; the other tasks POST `{}` to `/execute/<task>` and fail on non-2xx; no Task results or Workspaces. No prior reviews, other review sessions, generated policies, planner outputs, manuscript, `/tmp/tekton-accept-repo/results`, or full `live.py` were consulted. Local CLI guidance and command risk annotations were consulted solely to publish this document privately.

The predictions below assume a retry executes each selected task at most once, preserves serial ordering, includes the failed task and its remaining suffix, and stops at the first non-2xx. These retry scheduling rules are explicit assumptions: the service source itself does not define retryctl's selection or execution semantics. “Empty” means no earlier-task reselection; “declared” means the conservative relation derived below; “informed manual” means the minimum reselection using knowledge of the actual restored external world; “full restart” means all six pipeline tasks. A different submitted declaration must be evaluated separately rather than assigned these predictions automatically.

## Source semantics and captured states

Use `e` for the current epoch, and `c,a,s,q,d` for credential, artifact, signature, scanned, and deployed. The successful transitions are:

| Task | Guard | Assignment |
|---|---|---|
| warmup | None | No service state change |
| acquire | None | c := e |
| publish | c = e | a := e |
| sign | c = e and a = e | s := e |
| scan | s = e | q := e |
| deploy | q = e | d := e |

The one-shot injection precedes the task-specific guard/assignment. At the injected failure, it sets `injected=True`, returns 503, records an event, and leaves the task's output field unchanged. Thus retries from the captured state do not encounter another injected 503 unless a harness resets or changes that state.

| Source failure | Saved fields after failure at epoch 1 | Initially executed pipeline tasks |
|---|---|---:|
| publish | c=1; a=s=q=d=None; injected=True | 3 (warmup, acquire, publish) |
| sign | c=a=1; s=q=d=None; injected=True | 4 (warmup, acquire, publish, sign) |
| scan | c=a=s=1; q=d=None; injected=True | 5 (warmup, acquire, publish, sign, scan) |

Stable restoration keeps `e=1`. Rotated restoration changes `e=2` while retaining saved values of the fields above. Successful prior Tekton tasks are completed, even though their external products may now be stale. No task's guard reads Tekton status, a Task result, or a Workspace.

## Minimum earlier-task reselections

The failed task and not-yet-run suffix are mandatory in every successful retry and are excluded from the “earlier tasks” column.

| Failure | World | Minimum earlier tasks | Successful executed retry sequence | Pipeline-task count |
|---|---|---|---|---:|
| publish | stable epoch 1 | empty set | publish, sign, scan, deploy | 4 |
| sign | stable epoch 1 | empty set | sign, scan, deploy | 3 |
| scan | stable epoch 1 | empty set | scan, deploy | 2 |
| publish | rotated epoch 2 | acquire | acquire, publish, sign, scan, deploy | 5 |
| sign | rotated epoch 2 | acquire, publish | acquire, publish, sign, scan, deploy | 5 |
| scan | rotated epoch 2 | acquire, publish, sign | acquire, publish, sign, scan, deploy | 5 |

**Minimality proof.** In the stable world every required earlier field already equals the current epoch; zero earlier tasks is both sufficient and minimal. In the rotated publish case, only acquire can repair the credential guard. In the rotated sign case, only acquire and publish can repair its two guards, and publish requires acquire first. In the rotated scan case, only sign can repair the signature; executing sign requires current credential and artifact, so acquire and publish are necessary first. Warmup changes none of these fields. No smaller subset of the available earlier tasks repairs the relevant guards. This is minimality for the supplied task actions and states, not for arbitrary administrative state edits or alternative implementations.

## A single conservative retryWith relation

Define a relation from a failed task to earlier tasks that must accompany its retry:

```
retryWith(publish) = {acquire}
retryWith(sign)    = {acquire, publish}
retryWith(scan)    = {acquire, publish, sign}
```

This explicit relation works without assuming recursive expansion. If the retry mechanism recursively follows declarations, the compact generating relation is `publish -> acquire`, `sign -> publish`, `scan -> sign`; its transitive closure gives exactly the explicit sets above. A mechanism that reads only the failed task's direct list must use the explicit sets. The source alone cannot establish which representation retryctl implements.

The relation is the minimum fixed earlier-task selection per failure that succeeds in both specified worlds. It conservatively reexecutes some fresh work in the stable world. It is not a state-aware minimum there. It excludes warmup and supports the three requested failure locations; it makes no claim about arbitrary failures or later epoch changes during recovery.

## Expected executed-task counts and outcomes

Counts below cover the **retry attempt only**, include a task that executes and fails, and exclude tasks reused without executing and suffix tasks prevented from starting by a failure. They count pipeline tasks, including warmup when selected. These are analytical predictions, not measured outcomes.

| Failure | World | Empty | Declared relation above | Informed manual | Full restart |
|---|---|---:|---:|---:|---:|
| publish | stable | 4, success | 5, success | 4, success | 6, success |
| sign | stable | 3, success | 5, success | 3, success | 6, success |
| scan | stable | 2, success | 5, success | 2, success | 6, success |
| publish | rotated | 1, publish 409 | 5, success | 5, success | 6, success |
| sign | rotated | 1, sign 409 | 5, success | 5, success | 6, success |
| scan | rotated | 1, scan 409 | 5, success | 5, success | 6, success |

Empty rotated retries leave `deployed=None` and terminate before deploy. Their low executed-task counts represent failed recovery, so they cannot be treated as savings relative to successful recovery without reporting success as a separate outcome.

Across one retry in each of the six cases, empty executes 12 tasks and succeeds in 3/6 cases; declared executes 30 and succeeds in 6/6; informed manual executes 24 and succeeds in 6/6; full restart executes 36 and succeeds in 6/6. These six deliberately constructed cases are not a population sample or a basis for a general reliability rate.

Against full restart, the declared relation saves 1/6 executions per case, or 6/36 across the six cases. Manual saves 2/6, 3/6, and 4/6 in the stable publish, sign, and scan cases, respectively, and 1/6 in each rotated case: 12/36 overall. Declared uses 6 extra executions relative to manual, all in stable cases. These are counts of task invocations, not measured latency, CPU, money, or energy savings; warmup's constant-print task may cost much less than other tasks.

HTTP execution events differ from pipeline-task counts: full restart generates five `/execute` requests rather than six pipeline tasks. All other methods above skip warmup, so their HTTP counts equal their task counts. The original three captured runs generate 2, 3, and 4 HTTP events, respectively, and 3, 4, and 5 pipeline-task executions. Restored snapshots may carry their original events; retry counts require the post-restore event suffix, not the entire final event list. GET, reset, and restore are administrative operations and do not append task events in the supplied source.

If an end-to-end count is desired for a single source-plus-retry path, add 3, 4, or 5 to the appropriate retry count for publish, sign, or scan. Captured histories reused across strategies/worlds are shared histories, not additional physical executions; the reporting convention must make that reuse explicit.

## External final-state predicate

For restored epoch `e_expected` in {1,2}, define:

```
Recovered(state, e_expected) :=
    state.epoch == e_expected
    and state.credential == e_expected
    and state.artifact == e_expected
    and state.signature == e_expected
    and state.scanned == e_expected
    and state.deployed == e_expected
```

Read this state from the external service after recovery. It checks freshness across the credential/artifact/signature/scan/deployment chain independently of Tekton's success status or a planner's predicted selection. For these exact snapshots and fixed-epoch task transitions, successful deploy implies the earlier equalities, but checking all fields exposes accidental harness-state drift and states outside that restricted trajectory.

Separately check experimental integrity: `injected=True` remains set, `fail_at` is unchanged, epoch stays fixed throughout the retry, and the retry event suffix records the expected executed service tasks/status codes. Original injected 503 events remain legitimate in the retained source history; they must not make an otherwise successful retry fail an “all events are 2xx” check. A successful retry has all its own HTTP task events at 200. This predicate establishes only the simulator's modeled state, not real credential validity, cryptographic correctness, deployment availability, or production safety. Administrative restore can assign arbitrary state, so external-service separation alone does not establish independence of an experiment's oracle construction or harness.

## Ancestor and parallel validity

Every requested earlier task is a strict ancestor of its failed task in the serial pipeline. All explicit relation edges point backward in an acyclic task order; no self edge, descendant edge, or incomparable-task edge is needed. Serial execution in the original order satisfies the external guards after selection. Warmup is an ancestor but has no semantic repair value: graph ancestry is an admissibility check, not evidence of necessary external recovery.

The absence of Task results and Workspaces does not remove external dependencies. The service guards impose credential/artifact/signature freshness dependencies that are stronger than completed-task status alone.

This experiment contains no parallel branches and cannot empirically validate parallel selection, incomparable-task rules, races, or concurrent epoch rotation. If these service operations were made parallel, arbitrary ordering would be unsound: publish may run before acquire, sign before publish, scan before sign, or deploy before scan, causing 409s. The same task set is sufficient only with the required successful ordering. `HTTPServer` handles requests serially, but request arrival order from parallel clients would still need control. Claims about parallel validity require a separate DAG and execution evidence; no conclusion about a general retryctl parallel validator can be inferred from this source.

## Adjudication scope

The six state-conditioned minima, conservative relation, attempt counts, and predicate follow directly from the supplied deterministic source under the stated retry semantics. The evaluation can test whether a declared relation and actual execution repair the external state. It cannot, by itself, establish practitioner authoring accuracy, realistic policy discovery cost, broad failure coverage, production effectiveness, or a general external-recovery guarantee. Those conclusions require evidence beyond this source-only assessment.
