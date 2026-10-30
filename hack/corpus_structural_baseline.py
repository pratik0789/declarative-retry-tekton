#!/usr/bin/env python3
"""Offline structural cue/selection comparison; not a semantic recovery oracle."""
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import re
import zipfile

import yaml


RESULT = re.compile(r"\$\((?:tasks|finally)\.([^.\s)]+)\.results\.")


def strings(value):
    if isinstance(value, str):
        yield value
    elif isinstance(value, dict):
        for child in value.values():
            yield from strings(child)
    elif isinstance(value, list):
        for child in value:
            yield from strings(child)


def graph(spec):
    ordinary = spec.get("tasks", [])
    final = spec.get("finally", [])
    nodes = ordinary + final
    names = {n["name"] for n in nodes}
    assert len(names) == len(nodes), "duplicate task names"
    upstream = {name: set() for name in names}
    workspace = {name: set() for name in names}
    result_edges = set()
    for node in nodes:
        name = node["name"]
        upstream[name].update(node.get("runAfter", []))
        for text in strings(node):
            for producer in RESULT.findall(text):
                assert producer in names, (name, producer)
                upstream[name].add(producer)
                result_edges.add((producer, name))
        workspace[name].update(w["workspace"] for w in node.get("workspaces", [])
                               if "workspace" in w)
        assert upstream[name] <= names, (name, upstream[name] - names)
    for node in final:
        upstream[node["name"]].update(n["name"] for n in ordinary)
    shared_edges = {(a, b) for a in names for b in names
                    if a != b and workspace[a] & workspace[b]}
    return upstream, result_edges, shared_edges


def selection(upstream, seed, shared=()):
    """Rerun all graph predecessors, optionally closing shared Workspace peers."""
    chosen, pending = set(), [seed]
    while pending:
        task = pending.pop()
        if task in chosen:
            continue
        chosen.add(task)
        pending.extend(upstream[task] - chosen)
        pending.extend(a for a, b in shared if b == task and a not in chosen)
    return chosen


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--corpus", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--external-evidence", type=Path)
    args = parser.parse_args()
    packet = args.corpus / "frozen-source-packet.zip"
    adjudication = args.corpus / "adjudication.zip"
    with zipfile.ZipFile(packet) as archive:
        samples = json.loads(archive.read("selected-compositions.json"))
        index = {x["packet_path"]: x for x in
                 json.loads(archive.read("SOURCE_INDEX.json"))}
        graphs = {}
        for sample in samples:
            raw = archive.read(sample["composition_file"])
            assert hashlib.sha256(raw).hexdigest() == index[sample["composition_file"]]["sha256"]
            spec = list(yaml.safe_load_all(raw))[sample["document"]]["spec"]
            assert spec == sample["original_spec"]
            graphs[sample["pipeline_id"]] = graph(spec)
        eligibility = json.loads(archive.read("eligibility-ledger.json"))
    with zipfile.ZipFile(adjudication) as archive:
        ledger = json.loads(archive.read("adjudicated-relationships.json"))
    supported = [r for r in ledger if r["confidence"] == "Interface-supported"]
    pairs = []
    for record in supported:
        upstream, results, shared = graphs[record["pipeline_id"]]
        producer, consumer = record["producer"], record["consumer"]
        selected = selection(upstream, consumer)
        selected_ws = selection(upstream, consumer, shared)
        pairs.append({"pipeline_id": record["pipeline_id"], "producer": producer,
                      "consumer": consumer, "kind": record["kind"],
                      "coded_visibility": record["structural_visibility"],
                      "direct_result": (producer, consumer) in results,
                      "shared_workspace": (producer, consumer) in shared,
                      "producer_in_graph_ancestor_selection": producer in selected,
                      "producer_in_graph_workspace_selection": producer in selected_ws,
                      "graph_ancestor_selection": sorted(selected),
                      "graph_workspace_selection": sorted(selected_ws)})
    by_repo = Counter(r["pipeline_id"].split(":")[0] for r in supported)
    exclusions = {}
    for r in eligibility:
        repo = r["repo"]
        row = exclusions.setdefault(repo, {"candidates": 0, "selected": 0,
                                          "excluded": 0, "reason_counts": {}})
        row["candidates"] += 1
        if r["eligibility"] == "selected-source-bodies":
            row["selected"] += 1
        else:
            row["excluded"] += 1
            for reason in r.get("exclusion_reasons", []):
                row["reason_counts"][reason] = row["reason_counts"].get(reason, 0) + 1
    original_labels = []
    for filename in ["coder-a-original-labels.zip", "coder-b-original-labels.zip"]:
        with zipfile.ZipFile(args.corpus / filename) as archive:
            raw = json.loads(archive.read("relationships.json"))
        records = raw["records"] if isinstance(raw, dict) else raw
        labels = {}
        for r in records:
            key = (r["pipeline_id"], r["producer"], r["consumer"])
            labels[key] = labels.get(key, False) or r["confidence"] == "Interface-supported"
        original_labels.append(labels)
    a, b = original_labels
    common = set(a) & set(b)
    contingency = Counter((a[k], b[k]) for k in common)
    n = len(common)
    both, neither = contingency[True, True], contingency[False, False]
    a_only, b_only = contingency[True, False], contingency[False, True]
    observed = (both + neither) / n
    expected = ((both + a_only) * (both + b_only) +
                (neither + b_only) * (neither + a_only)) / n ** 2
    adjudicated_positive = {(r["pipeline_id"], r["producer"], r["consumer"])
                            for r in supported}
    result = {"scope": "Post-review offline comparison of source cues and conservative selections. "
                       "Supported pair records are conditional interfaces, not complete valid repair labels. "
                       "Lifecycle invalidators must not be interpreted as tasks that should rerun. "
                       "No operator study, native execution, error rate or policy-format advantage is inferred.",
              "corpus_sha256": hashlib.sha256(packet.read_bytes()).hexdigest(),
              "adjudication_sha256": hashlib.sha256(adjudication.read_bytes()).hexdigest(),
              "compositions": len(samples), "supported_pairs": len(pairs),
              "supported_by_repository": dict(sorted(by_repo.items())),
              "direct_result_pairs": sum(r["direct_result"] for r in pairs),
              "result_or_workspace_pairs": sum(r["direct_result"] or r["shared_workspace"] for r in pairs),
              "producer_in_graph_ancestor_selection": sum(r["producer_in_graph_ancestor_selection"] for r in pairs),
              "producer_in_graph_workspace_selection": sum(r["producer_in_graph_workspace_selection"] for r in pairs),
              "coding_agreement": {"n": n, "both_support": both, "neither_support": neither,
                                   "B_only_support": b_only, "A_only_support": a_only,
                                   "observed_agreement": observed, "expected_agreement": expected,
                                   "cohens_kappa": (observed - expected) / (1 - expected),
                                   "final_supported_set_equals_coder_B": adjudicated_positive ==
                                   {k for k in b if b[k]}},
              "canonical_candidate_exclusions": exclusions, "pairs": pairs}
    if args.external_evidence:
        pipeline = json.loads((args.external_evidence / "pipeline.json").read_text())
        upstream, _, _ = graph(pipeline["spec"])
        rows = []
        for failed in ["publish", "sign", "scan"]:
            for world in ["stable", "rotated"]:
                case = args.external_evidence / failed / world
                plan = json.loads((case / "declared" / "plan.json").read_text())
                continued = {r["name"] for r in plan["tasks"] if r["action"] == "continue"}
                declared = {r["name"] for r in plan["tasks"] if r["action"] != "inherit"}
                graph_executed = selection(upstream, failed) | continued
                assert graph_executed == set(upstream)
                native = json.loads((case / "full" / "taskruns.json").read_text())["items"]
                actual = {r["metadata"]["labels"]["tekton.dev/pipelineTask"] for r in native}
                assert actual == graph_executed and len(native) == len(actual)
                run = json.loads((case / "full" / "run.json").read_text())
                assert any(c["type"] == "Succeeded" and c["status"] == "True"
                           for c in run["status"]["conditions"])
                rows.append({"failed_task": failed, "world": world,
                             "ancestor_seeds": sorted(selection(upstream, failed)),
                             "continued_tasks": sorted(continued),
                             "graph_default_executed_tasks": sorted(graph_executed),
                             "declared_executed_tasks": sorted(declared),
                             "graph_default_taskruns": len(native),
                             "declared_taskruns": len(declared),
                             "native_evidence_status": "Same six-task action set as archived restart; not newly executed",
                             "raw_restart_taskruns": str((case / "full" / "taskruns.json").relative_to(args.external_evidence))})
        result["external_fixture_graph_default"] = rows
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps({k: v for k, v in result.items() if k not in
                      ["pairs", "canonical_candidate_exclusions", "external_fixture_graph_default"]}, indent=2))


if __name__ == "__main__":
    main()
