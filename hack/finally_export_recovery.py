#!/usr/bin/env python3
"""Verify a current-version recovery exporting a reused finally-task result."""

import argparse, hashlib, json, pathlib, subprocess, time, yaml

ap = argparse.ArgumentParser()
ap.add_argument("--context", default="kind-retry-paper")
ap.add_argument("--namespace", required=True)
ap.add_argument("--output", required=True)
ap.add_argument("--retryctl", default="/tmp/retryctl")
ap.add_argument("--resume", action="store_true", help="verify and continue an interrupted run in its original namespace")
ap.add_argument(
    "--prior-retryctl",
    help="Optional pre-fix binary to reproduce the dangling-reference manifest",
)
args = ap.parse_args()
root = pathlib.Path(__file__).resolve().parents[1]
ns = args.namespace
out = pathlib.Path(args.output)
out.mkdir(parents=True, exist_ok=args.resume)
k = ["kubectl", "--context", args.context, "-n", ns]


def call(args, data=None, check=True):
    read_only = data is None and args[0] == "kubectl" and any(x in args for x in ("get", "logs", "version"))
    for attempt in range(5):
        p = subprocess.run(args, input=data, text=True, capture_output=True, timeout=60)
        transient = any(x in p.stderr.lower() for x in ("request timed out", "timeout", "temporarily unavailable", "connection refused"))
        if p.returncode and read_only and transient and attempt < 4:
            time.sleep(attempt + 1)
            continue
        if check and p.returncode:
            raise RuntimeError(p.stderr)
        return p.stdout if check else p


def create(obj):
    if args.resume:
        found = call(k + ["get", obj["kind"], obj["metadata"]["name"], "--ignore-not-found", "-o", "json"])
        if found:
            existing = json.loads(found)
            def contains(actual, expected):
                if isinstance(expected, dict):
                    return isinstance(actual, dict) and all(key in actual and contains(actual[key], value) for key, value in expected.items())
                if isinstance(expected, list):
                    return isinstance(actual, list) and len(actual) == len(expected) and all(contains(a, e) for a, e in zip(actual, expected))
                return actual == expected
            assert contains(existing.get("spec", {}), obj.get("spec", {})), ("resume specification changed", obj["kind"], obj["metadata"]["name"])
            return existing
    return json.loads(call(k + ["create", "-f", "-", "-o", "json"], json.dumps(obj)))


def wait(name):
    end = time.monotonic() + 240
    while time.monotonic() < end:
        r = json.loads(call(k + ["get", "pipelinerun", name, "-o", "json"]))
        c = next(
            (
                c
                for c in r.get("status", {}).get("conditions", [])
                if c["type"] == "Succeeded"
            ),
            {},
        )
        if c.get("status") in ("True", "False"):
            return r, c["status"] == "True"
        time.sleep(1)
    raise RuntimeError("timeout " + name)


def save(name, obj):
    (out / name).write_text(json.dumps(obj, indent=2) + "\n")


def task(name, script, result=None, params=None):
    spec = {
        "steps": [
            {"name": "run", "image": "alpine:3.20", "script": "set -eu\n" + script}
        ]
    }
    if result:
        spec["results"] = [{"name": result, "type": "string"}]
    obj = {"name": name, "taskSpec": spec}
    if params:
        spec["params"] = [{"name": p["name"], "type": "string"} for p in params]
        obj["params"] = params
    return obj


metadata = {
        "namespace": ns,
        "context": args.context,
        "baseCommit": call(["git", "-C", str(root), "rev-parse", "HEAD"]).strip(),
        "harnessSHA256": hashlib.sha256(pathlib.Path(__file__).read_bytes()).hexdigest(),
        "generatorSourceSHA256": hashlib.sha256(
            (root / "internal/planner/create.go").read_bytes()
        ).hexdigest(),
        "fixedBinarySHA256": hashlib.sha256(
            pathlib.Path(args.retryctl).read_bytes()
        ).hexdigest(),
    }
if args.resume:
    saved = json.loads((out / "build-metadata.json").read_text())
    for field in ("namespace", "context", "generatorSourceSHA256", "fixedBinarySHA256"):
        assert saved[field] == metadata[field], ("resume metadata changed", field)
    save("resume-metadata.json", metadata)
else:
    save("build-metadata.json", metadata)
create({"apiVersion": "v1", "kind": "Namespace", "metadata": {"name": ns}})
save("kubernetes-version.json", json.loads(call(k + ["version", "-o", "json"])))
save(
    "tekton-version.json",
    json.loads(
        call(
            [
                "kubectl",
                "--context",
                args.context,
                "-n",
                "tekton-pipelines",
                "get",
                "configmap",
                "pipelines-info",
                "-o",
                "json",
            ]
        )
    ),
)
p = {
    "apiVersion": "tekton.dev/v1",
    "kind": "Pipeline",
    "metadata": {"name": "finally-export"},
    "spec": {
        "tasks": [
            task(
                "build", 'printf "%s" source-image > "$(results.image.path)"\n', "image"
            )
        ],
        "finally": [
            task("audit", 'printf "%s" checked > "$(results.note.path)"\n', "note"),
            task(
                "notify",
                'case "$(context.taskRun.name)" in *-source-*) echo injected-failure; exit 1;; esac\nprintf "%s" "$(params.image)" > "$(results.receipt.path)"\n',
                "receipt",
                [{"name": "image", "value": "$(tasks.build.results.image)"}],
            ),
        ],
        "results": [{"name": "audit-note", "value": "$(finally.audit.results.note)"}],
    },
}
p = create(p)
save("pipeline.json", p)
create(
    {
        "apiVersion": "tekton.dev/v1",
        "kind": "PipelineRun",
        "metadata": {"name": "finally-export-source"},
        "spec": {"pipelineRef": {"name": "finally-export"}},
    }
)
r, ok = wait("finally-export-source")
assert not ok
save("source-run.json", r)
trs = json.loads(
    call(
        k
        + [
            "get",
            "taskruns",
            "-l",
            "tekton.dev/pipelineRun=finally-export-source",
            "-o",
            "json",
        ]
    )
)
save("source-taskruns.json", trs)
by = {t["metadata"]["labels"]["tekton.dev/pipelineTask"]: t for t in trs["items"]}
assert {
    n: next(c["status"] for c in t["status"]["conditions"] if c["type"] == "Succeeded")
    for n, t in by.items()
} == {"build": "True", "audit": "True", "notify": "False"}
assert (
    next(x["value"] for x in by["audit"]["status"]["results"] if x["name"] == "note")
    == "checked"
)
policy = {
    "apiVersion": "retry.tekton.dev/v1alpha1",
    "kind": "PipelineRetryPolicy",
    "metadata": {"name": "finally-export"},
    "spec": {"pipelineRef": "finally-export", "tasks": {}},
}
save("policy.json", policy)
a = [
    "--pipeline",
    str(out / "pipeline.json"),
    "--run",
    str(out / "source-run.json"),
    "--taskruns",
    str(out / "source-taskruns.json"),
    "--policy",
    str(out / "policy.json"),
]
validation = None
if args.prior_retryctl:
    old = call([args.prior_retryctl, "create"] + a)
    (out / "before-fix.yaml").write_text(old)
    oldspec = yaml.safe_load(old)["spec"]["pipelineSpec"]
    assert oldspec["results"][0]["value"] == "$(finally.audit.results.note)"
    assert [t["name"] for t in oldspec["finally"]] == ["notify"]
    validation = call(
        k + ["create", "--dry-run=server", "-f", "-", "-o", "json"], old, check=False
    )
    save(
        "before-fix-validation.json",
        {
            "returncode": validation.returncode,
            "stdout": validation.stdout,
            "stderr": validation.stderr,
        },
    )
new = call([args.retryctl, "create"] + a)
(out / "generated.yaml").write_text(new)
manifest = yaml.safe_load(new)
publishers = manifest["metadata"]["annotations"][
    "retry.tekton.dev/result-publishers"
].split(",")
assert len(publishers) == 1
assert (
    manifest["spec"]["pipelineSpec"]["results"][0]["value"]
    == "$(tasks." + publishers[0] + ".results.value-0)"
)
assert [t["name"] for t in manifest["spec"]["pipelineSpec"]["finally"]] == ["notify"]
save("plan.json", json.loads(call([args.retryctl, "plan"] + a)))
manifest["metadata"].pop("generateName", None)
manifest["metadata"]["name"] = "finally-export-recovery"
save("submitted.json", manifest)
create(manifest)
recovery, ok = wait("finally-export-recovery")
save("recovery-run.json", recovery)
assert ok
newtrs = json.loads(
    call(
        k
        + [
            "get",
            "taskruns",
            "-l",
            "tekton.dev/pipelineRun=finally-export-recovery",
            "-o",
            "json",
        ]
    )
)
save("recovery-taskruns.json", newtrs)
assert len(newtrs["items"]) == 2
recovered_by_name = {
    t["metadata"]["labels"]["tekton.dev/pipelineTask"]: t for t in newtrs["items"]
}
assert set(recovered_by_name) == {"notify", publishers[0]}
assert (
    next(
        x["value"]
        for x in recovered_by_name["notify"]["status"]["results"]
        if x["name"] == "receipt"
    )
    == "source-image"
)
assert (
    next(
        x["value"]
        for x in recovered_by_name[publishers[0]]["status"]["results"]
        if x["name"] == "value-0"
    )
    == "checked"
)
assert (
    next(x["value"] for x in recovery["status"]["results"] if x["name"] == "audit-note")
    == "checked"
)
assert "retry.tekton.dev/finally-anchor" not in manifest["metadata"]["annotations"]
summary = {
    "sourceFailedOnlyInNotify": True,
    "currentStrictGuardAcceptedNativeCapturedInputs": True,
    "oldGeneratorLeftRemovedFinallyResultReference": True
    if validation is not None
    else None,
    "oldManifestServerDryRunAccepted": validation.returncode == 0
    if validation is not None
    else None,
    "recoverySucceeded": True,
    "executedOriginalTasks": ["notify"],
    "resultPublishers": publishers,
    "recoveryTaskRuns": 2,
    "auditReused": True,
    "exportedAuditNote": "checked",
    "receipt": "source-image",
    "auditWorkNotRepeated": True,
    "publisherEmitsSavedString": True,
    "scope": "One synthetic current-version composition regression, not a performance or production-effectiveness study.",
}
save("summary.json", summary)
print(json.dumps({"evidence": str(out), "namespace": ns, **summary}))
