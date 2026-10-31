#!/usr/bin/env python3
"""Optional structural advisory for reuse between tasks that execute again.

This does not modify a recovery plan or infer the validity of external effects.
It inspects runAfter, result references, and Tekton's ordinary/final phase.
Usage: python3 hack/check_reuse_convexity.py --pipeline pipeline.yaml --plan plan.json
"""
import argparse
import json
import re
from pathlib import Path

import yaml


RESULT = re.compile(r"\$\((?:tasks|finally)\.([A-Za-z0-9_-]+)\.results\.[^)]+\)")


def inspect_reuse(pipeline, plan):
    if plan.get("decision") != "recover":
        return []
    if plan.get("pipeline") != pipeline.get("metadata", {}).get("name"):
        raise ValueError("plan and Pipeline names differ")
    spec = pipeline["spec"]
    ordinary = spec.get("tasks", [])
    final = spec.get("finally", [])
    tasks = ordinary + final
    names = [t["name"] for t in tasks]
    if len(names) != len(set(names)):
        raise ValueError("duplicate Pipeline task name")
    index = {n: i for i, n in enumerate(names)}
    reach = [[False] * len(names) for _ in names]
    for task in tasks:
        parents = set(task.get("runAfter", []))
        parents.update(RESULT.findall(json.dumps(task)))
        if task in final:
            parents.update(t["name"] for t in ordinary)
        for parent in parents:
            if parent not in index:
                raise ValueError(f"unresolved dependency {parent!r}")
            reach[index[parent]][index[task["name"]]] = True
    for k in range(len(names)):
        for i in range(len(names)):
            for j in range(len(names)):
                reach[i][j] |= reach[i][k] and reach[k][j]
    if any(reach[i][i] for i in range(len(names))):
        raise ValueError("cyclic execution dependencies")
    entries = plan.get("tasks", [])
    actions = {t["name"]: t["action"] for t in entries}
    if len(entries) != len(actions) or set(actions) != set(names):
        raise ValueError("plan does not partition the Pipeline tasks")
    if any(a not in {"rerun", "continue", "inherit"} for a in actions.values()):
        raise ValueError("unknown plan disposition")
    execute = {n for n, a in actions.items() if a != "inherit"}
    warnings = []
    for name in sorted(names):
        if actions[name] != "inherit":
            continue
        ancestors = sorted(n for n in execute if reach[index[n]][index[name]])
        descendants = sorted(n for n in execute if reach[index[name]][index[n]])
        if ancestors and descendants:
            warnings.append({
                "code": "non-convex-reuse",
                "task": name,
                "executedAncestors": ancestors,
                "executedDescendants": descendants,
                "message": (
                    "Reused work lies between tasks scheduled to execute. "
                    "Check whether rerunning its ancestors invalidates its effects "
                    "or certification; explicitly select the affected tasks if needed. "
                    "Execution order alone does not establish invalidity."
                ),
            })
    return warnings


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pipeline", required=True)
    parser.add_argument("--plan", required=True)
    args = parser.parse_args()
    pipeline = yaml.safe_load(Path(args.pipeline).read_text())
    plan = json.loads(Path(args.plan).read_text())
    print(json.dumps({
        "advisories": inspect_reuse(pipeline, plan),
        "scope": "Structural check only; no external-effect validation, plan mutation, or refusal.",
    }, indent=2))


if __name__ == "__main__":
    main()
