# Recovery contract: observable boundary, compositional selection, and DAG projection

## Scope and definitions

A fixed resolved Pipeline has a finite directed acyclic execution graph G=(V,E), with a->b when a is required to finish before b. References to an ordinary task's result count as execution dependencies, including references in parameters, conditions and embedded task bodies. The ordinary-to-final phase barrier is an additional ordering constraint. The proof below treats ordinary nodes; final nodes preserve their native phase, and a separate native experiment covers final-task admission.

A source observation O contains the Pipeline specification, source PipelineRun, complete terminal source TaskRuns and their string results, and bindings. The source PipelineRun's Succeeded condition must be False. Every observed child must have a terminal Succeeded/Failed status. Observable terminal status does not prove that remote work stopped, that a timed-out process has no surviving effects, or that snapshots are atomic. Author/operator assumptions still exclude concurrent effects and mutation of reused values. The CLI checks the observable status boundary, not global quiescence.

Let R be the author-declared recovery relation, directed from a selected task to a task it requires to repeat. Validity requires that the closure of each root contains no execution descendant of that root. Let S be the failed and explicitly selected seeds; C_R(S) is its least R-closed superset. Let H be the successful source tasks outside C_R(S), and K=V\H. K includes both rerun tasks and tasks not reached by the source.

## Proposition 1: an observable planner cannot always be both safe and reuse-maximal

There exist two external states e0,e1 with the same O for which the smallest sufficient rerun sets differ. Consider acquire->publish->sign->scan->deploy. None exports a Task result or mounts a Workspace; each exchanges state with a remote service. An injected transient failure stops scan. In e0, the service still accepts the saved credential, artifact and signature. Rerunning scan alone, then continuing deploy, is sufficient. In e1 the service rotates its epoch after source capture; the saved credential, artifact and signature refer to the old epoch. Recovering requires acquire, publish and sign to repeat before scan.

The Pipeline and all captured source Tekton objects are identical in both worlds. For any deterministic planner f with input O alone, f(O) is identical in both. Selecting only scan is insufficient in e1. Selecting all four is safe for the fixture in both, but is not reuse-maximal in e0. Therefore no f based on O alone can always satisfy both properties on this class. A randomized planner cannot satisfy both with certainty either. Full restart is a conservative fallback and is not excluded by this proposition. A richer external oracle changes the premise.

This is an information boundary, not a new impossibility result for workflow engines. It makes precise why a warning-free plan is not evidence of semantic safety. A conservative policy can encode the application's tolerance of both worlds. It cannot discover the current epoch, prove the policy is complete, or guarantee globally optimal reuse.

## Proposition 2: composition of recovery selections

For fixed R, C_R(S1 union S2) = C_R(S1) union C_R(S2).

Proof. Any node reachable from a seed in S1 union S2 is reachable from a seed in S1 or S2. Conversely, every node reachable from either seed set is reachable from their union. Thus both sides are exactly the same reachable-node set. This also proves extensivity, idempotence and monotonicity. These are standard reachability properties, not claimed as new graph theory.

Consequence. A validated relation supplies one policy-compliant selection function for multiple failed/explicitly selected tasks; it does not require enumerating all 2^|V| seed sets. An equally expressive persisted manual rule can implement exactly the same function. No authoring-time, error-rate or representational superiority over that baseline follows. The prototype contribution is an executable common contract with validation, transformation and provenance, whose limits are separately tested.

## Proposition 3: exact ordering projection through inherited tasks

Define a projected graph G_K on K: retain any original edge with both endpoints in K; for each original path whose endpoints are in K and whose internal nodes all lie in H, add an endpoint edge. Then, for all a,b in K, a reaches b in G_K if and only if a reaches b in G.

Proof (soundness). Every G_K edge is an original edge or stands for an original path. Concatenating these paths gives an original walk; because G is acyclic it establishes reachability and cannot introduce a cycle. Proof (completeness). Split any original a-to-b path at each retained node. Each segment has inherited internal nodes or is a direct retained edge. Each segment therefore has a corresponding edge in G_K; concatenating them yields a projected path. Nearest-retained predecessor traversal implements exactly those segments. Redundant edges preserve reachability.

Implementation qualification. Native result dependencies between retained nodes remain in their result expressions. Dependencies whose producers are inherited are replaced with saved string values, and corresponding ordering segments are reconstructed as runAfter edges. Missing/non-string inherited values or remaining unsupported references refuse generation. The proof establishes scheduling reachability, not extensional equivalence of mutable effects or exact recovery output. The exhaustive checker uses an independent Boolean Floyd-Warshall oracle on 1,024 forward DAGs with five labeled vertices and all 31 nonempty retained subsets: 31,744 transformations. Edge encodings mix runAfter, parameter-result, and when-result references. These are abstract partitions, including partitions not realizable as failed execution traces, and are not 31,744 native cluster tests. All forward DAGs on a fixed topological ordering are covered; relabeling represents any five-node DAG. Larger graphs follow the proof, while arbitrary Tekton language features remain outside the fixture check.

## Combined contract

For a fixed valid policy and admissible source, selection supplies the least policy-compliant rerun set and generation preserves original ordering among retained ordinary tasks, while inherited string uses are frozen to their source values. If author assumptions establish that inherited values/effects remain valid under the recovery execution, this is a structurally faithful way to realize the policy. Neither selection nor graph projection establishes those semantic assumptions. Refusing nonterminal sources and aligning warning extraction close two observable implementation gaps; neither converts advisories into semantic guarantees.

## Empirical link

The external-service experiment uses three distinct failed tasks (publish, sign, scan), each captured once. Exactly the same source objects are passed to each arm in two independently restored external worlds. A fixed policy is applied without per-failure edits. Arms are empty selection, declared closure, policy-informed manual selection using the same generator, and full restart. The oracle records HTTP effects and final external epochs without access to retryctl plans or Tekton APIs. Its semantics, tasks and policies are project-authored, so this is an oracle separated from planner implementation, not an independent policy-author or production study. One execution per failure/world/arm gives 24 native arms; no population success probability or authoring efficiency is inferred.

## Bounded semantic minimality and observable information

For the specified service transitions, each epoch-bearing field has one producer. After rotation, publish needs acquire; sign needs acquire and publish; scan needs acquire, publish and sign. These sets are necessary by the guards and unique producers, and sufficient in the fixed serial execution. They are also the smallest fixed earlier selections that work in both admitted worlds. In the stable world none is necessary; a state-aware oracle can do better. This proves robust minimality only for the enumerated fixture semantics and two worlds, not an arbitrary external effect model. The graph provides admissible ordering, the declared relation supplies the conservative repair prerequisites, and the native comparisons test their composition.

The live manual arm is policy-informed and receives the same conservative requirements as the declared arm. A separate state-aware minimum comparator is reconstructed from the existing stable empty and rotated manual runs. It uses 24 TaskRuns across six cases versus 30 for the conservative policy, with zero additional trials. Source-only automated adjudicators derived these minima before inspecting outcomes. The six extra conservative executions in stable worlds are the cost of declining to observe the external epoch. This is an explicit information/work tradeoff, not a measured authoring or performance advantage.
