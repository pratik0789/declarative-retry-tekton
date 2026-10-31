# Value study for declarative Tekton recovery

**Status: prospective draft for feasibility review. No study, participant session, failure replay, or result is reported here.**

## Decision and scope

Functional agreement with informed manual selection establishes policy fidelity, but does not measure the value of storing recovery knowledge. This prospective protocol specifies the additional empirical question.

Use **real failure histories** if usable captures and recorded recovery choices are available. Otherwise use a **controlled human authoring study**. A static public-source corpus is a fallback for characterizing recovery obligations; it does not establish incident frequency or human benefit. Automated agents must be described as automated coders, never participants or independent human authors.

The research question is: *Does storing and reapplying recovery knowledge reduce observed recovery mistakes or total work across repeated incidents, and what costs and applicability limits does it introduce?* The question is about an implemented workflow and its users. An equally expressive persisted rule may match every behavior and benefit; that is a required control, not a failure to be removed from the study.

Before execution, freeze the final implementation commit/binary hash, all source snapshots, cases, oracle definitions, comparator implementations, assignments, exclusions and analysis plan. Publish that timestamped packet. Freezing a packet after seeing its outcomes is not preregistration.

## Shared controls and outcomes

### Equal information and expressive controls

All arms receive the same resolved Task/Pipeline definitions, failure/status captures, available state observations, domain-contract facts and permissible actions. No arm receives the answer selection, later outcome, or hidden domain fact another arm lacks. All use the same reduced-run generator and admission/warning behavior, so result inheritance or generator bugs do not confound selection value.

| Arm | Workflow | Role in interpretation |
|---|---|---|
| M: per-incident selection | Supply explicit task selections to the common generator. Operators may retain notes, reuse earlier work and write scripts; record that activity. | Represents per-incident practice with equally available knowledge, rather than a deliberately forgetful baseline. |
| P: generic persisted rule | Store an equally expressive relation/refusal rule and apply a supplied compiler. It supports the same closure, admission, ordering and content-history checks. | Controls for persistence and automation. It must be able to reproduce C's plans; generated table size is not a benefit metric. |
| C: declarative policy | Author/revise a retryctl policy and apply it through the common generator. | Tests the implemented workflow. Comparison with P limits any claim unique to this policy format. |
| R: restart | Restart all applicable tasks from the same admissible source/state. | Work-cost reference; not an equally informed manual control. |

P should be generated from the same neutral fact ledger and independently checked for equivalence on admissible inputs. Do not withhold validation, transitivity or history from it. Differences in optional documentation, user interface or defaults must be named and measured; they are not expressiveness differences. If M operators build reusable automation, keep their original assigned arm and report their strategy.

### Outcome oracle

Define correctness from existing application/service semantics before the policies or human decisions are scored. The oracle reads actual task effects, identifiers, coverage and final payloads; it does not read the selected task set, policy names, plan labels or expected repair membership. Use exact source/image/resource identities and immutable input fixtures. An empty successful result is not proof all required work happened.

Examples of sufficient checks depend on the selected application: an emitted OCI object contains the requested application's bytes and configuration; every required component was actually scanned against the specified scanner state and its saved payload matches; or a restored resource has the required incarnation-specific initialized state. A model that merely encodes "tasks X and Y must rerun" is not an independent oracle. If full effects cannot be established, label the endpoint a necessary-precondition check and keep it out of the full-validity denominator.

The oracle curator must be separate from the policy author and blinded to arm assignments. Record who wrote it, source citations, prompts/model versions for any automated assistance, and exactly when the inputs and outputs were visible. Prefer a non-project engineer authoring at least one policy from upstream contracts. Human scoring by agreement alone is insufficient: disagreements need the actual Task/API evidence, and unresolved cases remain unknown.

### Measurements and units

Primary human endpoints are (1) **unsafe-action proportion** and (2) **total active work** per participant/block. Total work includes training attributable to the workflow, initial rule authoring, policy/rule updates, investigation, selection, verification and correction. Do not omit setup while claiming amortized savings. A safe conservative restart is not an omission error; count its extra work separately.

Separate unsafe omissions, unnecessary repeats, incorrect refusals, genuinely required refusals and unscorable cases. Report the first submitted decision and the final corrected decision. Timeout/abandonment remains in its assigned arm. Count task-minutes and elapsed recovery time separately: summing concurrent TaskRun durations is aggregate work, not lead time.

An optional reconstruction endpoint is a second operator's correctness/time reconstructing an earlier recovery from its retained packet. Give M and P equivalent opportunities to retain source hashes, selections, rule content and notes. Default unauthenticated annotations are inspectable records, not tamper-proof audit evidence.

## Route A: retrospective on real failure histories

### Minimum data and eligibility

Use one identifiable deployed Tekton system and a predeclared consecutive time window. A useful descriptive pilot needs approximately **30 eligible failed PipelineRuns across at least three distinct pipelines**, including repeated incidents; that is an operational target, not a power calculation or acceptance guarantee. Freeze the window and enumerate **all** failures before selection. Report the complete screening ledger, including exclusions and missing captures. Do not extend the window solely until favorable cases appear.

Each eligible incident needs:

1. Terminal failed PipelineRun and observed TaskRuns with timestamps and attempts; the actual resolved Pipeline/Task definitions and effective parameters.
2. Contemporaneous artifact/resource identifiers and enough retained state to determine what a later consumer could reuse. A saved scalar result alone is insufficient.
3. Actual manual recovery choice, its source-time information and subsequent run/output evidence when comparing with human practice. If unavailable, the incident can support work/obligation characterization but not a manual-error claim.
4. An external output or coverage predicate and a supported-source interpretation under the frozen retryctl build. Unsupported definitions and mismatches must be reported rather than edited to force admission.
5. Pre-incident domain documentation/code sufficient for a policy author to work without seeing later outcomes. Preserve the source chronology; policies cannot silently use hindsight-only facts.

A non-project author writes one policy per pipeline using the outcome-blinded packet; a second coder checks the underlying contracts independently. Lock the policies, then reveal the observed recovery/outcome files. Recompute all four arms' dispositions. Do not swap definitions across revisions to avoid source admission.

### Required analyses

- At incident level: failed roots, selected/reused task counts, changed selection relative to failed-task-only and restart, warnings, refusal reason, policy-induced extra reruns, and oracle validity/unknown.
- At pipeline level: how many incidents use the same declaration unchanged; which genuine upstream revisions require policy maintenance; policy-authoring/update effort if actually measured.
- Aggregate original successful TaskRun durations for work potentially retained. State that this is a **counterfactual original-task-work estimate**, not observed latency saved by a new run. Parallelism, caches, admission overhead and changed side effects prevent equating these quantities.
- If disposable replay is feasible, execute admitted source-coherent arms in matched frozen states with randomized order; record native status and external output separately. A single execution per cell is descriptive. Do not label a replay an original production recovery.
- Compare the locked policy with **recorded actual manual choices** only when their available information is matched. Report expertise, changed state and hindsight as confounders. A newly expert-authored policy beating an uninformed historical choice does not prove an encoding advantage.

Use the pipeline, not the incident, as a dependence unit for uncertainty summaries; with only three pipelines, show per-pipeline results and treat pooled intervals cautiously. No population/prevalence claim outside the sampled deployment/window follows. Unknown semantic validity must remain visible in numerator/denominator reporting. The effect of an observed unsafe manual selection may be reported as a counterfactual candidate avoided error only when the locked oracle actually establishes it.

**Can prepare now:** capture schema, screening ledger, outcome-blind packet layout, generic persisted-rule control and analysis definitions. **Cannot replace with synthetic data:** original captures, manual choices, resource-state evidence or observed recovery outcomes.

## Route B: controlled human authoring study

### Participants and case bank

Prepare a descriptive pilot of **12 engineers**, including engineers outside the paper project, with documented CI/Tekton experience. Do not substitute 12 agent runs for 12 people. Report recruitment basis, experience and all exclusions; disclose participant study information and logging. Study contacts require the user's explicit outreach authorization. This protocol itself makes no contact.

Build three source-resolved public/deployed compositions with materially different effect contracts. Each block contains three repeated failure contexts plus one genuine upstream revision requiring a maintenance decision: **four decisions per arm, 12 decisions per participant**. Contexts must come from genuine failure histories or source/API-derived executable tasks selected before checking treatment outcomes. Project-authored failure injection is permissible only when transparently labeled and its oracle is not a restatement of the policy. Do not invent a revision to create a desired maintenance result.

Use distinct matched compositions across arms to limit answer memorization. Freeze a balanced assignment of arm-to-composition and block order; each arm must occur equally at each position and with each composition. Stratify experience where feasible and lock the random seed before recruiting. Pilot the instructions/interface with a separate person whose results are excluded from inference; if materials change, record the change and freeze the final packet before the study.

### Procedure

1. Give equal core training, source facts and access to references. Test comprehension; keep failures and retries in the record.
2. Begin logging before workflow-specific training/authoring. M begins selecting per incident; P and C first author their respective rules. Operators may ask the same factual clarification channel in every arm; log supplied information.
3. Present the first three contexts successively. Collect the first selection/refusal, confidence and explanation, then apply it to the common executor/oracle. Provide matched feedback. Record corrections and all active time.
4. Reveal the genuine upstream revision equally in every arm. Ask for the necessary maintenance and next recovery, allowing exactly the same source information and tools. Record update/verification time and stale-rule errors.
5. Optionally assign blinded reconstruction of another operator's retained incident packet, with comparable metadata-retention facilities in all arms.

### Analysis and success interpretation

Predeclare C-minus-M and C-minus-P differences in unsafe-action proportion and total active work. Report participant-level paired results, distributions, and uncertainty with the participant as the cluster; incident rows are not independent people. With 12 participants, report this as a small exploratory/descriptive study, without claiming population effects or treating a non-significant result as equivalence. Report setup and incident-only time both, and show whether savings amortize within the observed four incidents. Separate oracle-correct refusal from failure to produce any useful action.

A C/P match with both improving on M supports the practical value of **persisting recovery knowledge**, not unique superiority of retryWith syntax or a novel closure algorithm. C/P/M all matching with equal total effort is a negative/inconclusive value result and must be reported. An apparent C advantage that comes from unmatched facts, withheld P features, different generators, hidden baseline setup or a policy-encoded oracle does not answer the research question. Any update-related unsafe regression or excessive conservative work must be reported alongside benefits.

**Can prepare now:** cases' source requirements, task instructions, assignments, common executor/scorer, logging schema and compiler control. **Data still required:** actual engineers and complete executable cases/oracles. Do not write error/time numbers before sessions happen.

## Route C: public-source corpus fallback

Reviewer 2 explicitly accepts a corpus of at least 30 public Tekton pipelines coding warning-invisible recovery obligations. Reviews 1 and 3 demand real incidents or human authoring evidence; **this fallback cannot be promised to close all three gates**.

Freeze a source list using a stated eligibility/search rule, not favorable outcomes. Resolve complete definitions; count compositions, not Task slots or cosmetic variants. Two coders independently code consumed state, lifetime/ownership, initialization effects, source contract evidence, failure context, current warning observability and expressibility under the actual policy restriction. Publish disagreement and unknown categories; an automated second coder is acceptable only for a transparently automated source analysis. Standard closure or typed-file-availability examples alone add limited research value.

Report proportions only within the sampled source set; no deployment-frequency, operator-error or actual time-saving claim follows. Externally grounded execution of a subset can check the coded contracts but is still not an operator study. Existing inspected allocation/auth/signing candidates in earlier source-screening work informed this proposal and cannot retrospectively become held-out tests. Their unresolved historical runtime compatibility must remain excluded from complete old/new validity claims.

## Reviewable evidence packet and claim gate

The packet should contain the frozen source/capture index and hashes; complete screening ledger; oracle specification and provenance; policy-author packets with outcomes withheld; M/P/C interfaces; common generator/binary hashes; assignments; empty record schema; and analysis plan. After execution append raw outcomes, exact clocks/actions, all negative/unscorable cases and reproduction commands.

Before claiming the value gap is addressed, check:

- At least one endpoint measures a real practical outcome on eligible histories or people, beyond plan equivalence and generated row counts.
- The policies/oracle/outcomes are separated in authorship and chronology; domain facts are equal across controls.
- Setup, maintenance, verification, conservative work and unsafe/failed cases are counted.
- The generic persisted control is allowed to match; claims name the measured causal contrast.
- The revised paper reports the actual study scope and consolidates the current build. Existing demonstrations are not renamed as field evidence or human validation.

A protocol, static corpus or favorable pilot does not guarantee a venue recommendation. Freeze the revised paper/evidence only after actual results exist, then request a new independent review without specifying its verdict.

