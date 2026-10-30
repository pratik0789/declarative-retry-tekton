# Public GitHub issue captures: frozen feasibility screen

## Result and scope

**No usable completed failed PipelineRun case packet was established from the fixed candidate set.** Twenty-two unique issue bodies were inspected, under the cap of sixty. They were selected by twelve queries fixed before bodies were opened. This is purposeful feasibility discovery, not a representative failure cohort or evidence of a population frequency, policy benefit, operator error rate or venue outcome.

The first frozen file specifies four queries per repository—tektoncd/pipeline, tektoncd/catalog and konflux-ci/build-definitions—combining PipelineRun with pipelineSpec, taskRuns, completionTime, or conditions/Failed. Each query takes up to five results sorted by creation time descending. The complete search responses were saved, then an exact candidate-URL ledger was frozen before bodies were inspected. Duplicate URLs were read once. Queries and selection were not changed after seeing results.

The accompanying ZIP contains the fixed protocol, candidate ledger, raw query responses, each issue's original public JSON/body with hashes, and a frozen per-issue eligibility/exclusion ledger. No native resource object, field or policy label has been reconstructed.

## Genuine operational leads, with explicit missing fields

### Parent remains failed after its child recovers

https://github.com/tektoncd/pipeline/issues/9467

The issue reports actual TaskRun and PipelineRun reconciliation logs for a Swift-operator run in openstack-tenant. The reported parent name is `swift-operator-openstack-18-0-on-pull-request-nvvsc`; the child build-container TaskRun is also named in the original body. On 2026-01-16 the child briefly reports Failed after init-container exit 137; at about 10:01:20 the parent reports Failed and fourteen later Tasks skipped. The same child returns to Unknown/Running and reports Succeeded at 10:05:30. These are verbatim reported controller observations, not a captured resource snapshot or a new execution.

This is a credible incident lead for the distinction between terminal parent status and live child activity. It **does not** supply:

- the original raw PipelineRun and TaskRun objects, including source UID/spec/status;
- a resolved/versioned Pipeline and relevant Task definitions and runtime versions;
- complete sibling TaskRun status/results/times or retained external output facts;
- an actual manual task-selection repair or a source-linked recovery PipelineRun;
- an output-validity oracle or evidence that a proposed rerun would succeed.

The reported later child success is **within the existing TaskRun**, not proof of an externally chosen recovery PipelineRun. No avoidable time, correct selection, policy necessity or practical contract benefit is inferred. This packet remains ineligible for the retrospective.

### Failed PipelineRun status/Event with missing workflow

https://github.com/tektoncd/pipeline/issues/9002

The body provides a terminal status fragment (`RequiredWorkspaceMarkedOptional`), a dated native Event with run name/namespace/UID, and Tekton v1.0.0. It does not provide the original Pipeline/Task spec or TaskRun state. The failure is validation of an optional/required Workspace mismatch before the application executes, so it cannot establish reusable successful application work or recovery task selection. The Event text was retained as originally published; it was not converted into a fabricated PipelineRun object.

### Task-version incompatibility and a manual configuration attempt

https://github.com/tektoncd/catalog/issues/1044

The issue records git-clone 0.7 versus 0.8 behavior and an attempted `deleteExisting=false` workaround followed by an actual git-init permission-error log. The linked ZIP is described as reproduction manifests. This is genuine public evidence of a version/permission issue and a manual **configuration** attempt, but it lacks completed source PipelineRun/TaskRun captures and is not a recorded manual **task selection** repair. No ZIP execution was performed.

https://github.com/tektoncd/catalog/issues/531 similarly reports a git-clone reinstall incompatibility with a Task invocation fragment and a pinned external example. It lacks the original failed PR/TR state. Neither issue qualifies merely because a template/Task version and error log are public.

## Other exclusions

The ledger gives one reason for every frozen candidate URL. The remaining bodies describe:

- deliberately constructed enum, timeout, admission, missing-secret or resolver stress reproductions;
- nonterminal Running/Pending captures;
- TaskRun-only submissions or pod/container output without the parent workflow;
- aggregate retention/event/controller observations without individual captured runs;
- sample-definition mismatches or feature requests.

The tool does not infer a completed failed PipelineRun from a Pod condition `False`, a Task's nonzero exit, an issue marked closed, a code fix, or a later Task success. It also does not equate a reproduction manifest with the immutable/resolved source of a naturally occurring application failure.

## What is frozen and what remains missing

`FROZEN_SCREEN_DECISIONS.json` contains **zero eligible captures** and all twenty-two excluded candidates, plus exact body/source hashes. The nearest genuine incident is #9467; it needs the deployment's archived raw source PR/TR captures, resolved definitions, full Task/state history and actual recovery linkage. Without those records, it may support a narrowly attributed motivating incident discussion, but it cannot clear the requested retrospective/value-study gap.

No source/manuscript/cluster files were edited; no tasks, models or upstream project code were executed; no authors or humans were contacted. Public API reads and local source/index serialization were the only operations. No further queries were added after the frozen searches.
