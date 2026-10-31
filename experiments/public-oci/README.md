# Public-task OCI application recovery case

This is a bounded authored integration of pinned public Tekton Catalog tasks,
Skaffold application/Dockerfile, and Docker Distribution API semantics. It is
neither a production incident nor a practitioner study. `protocol.md` defines
profiles/worlds/arms and calibration changes. All application and Task scripts
remain public-source bytes. `model_audit.py` checks the recorded adaptations and
identical execution reachability of the two profiles.

## Offline audit

Use Python 3 with PyYAML; no cluster, Docker, or network is needed:

```sh
python3 experiments/public-oci/model_audit.py
python3 experiments/public-oci/audit.py
```

The content checker independently reconstructs the delivered `/app` from actual
content-addressed OCI layers using reverse traversal, compares binary/config,
and checks native run/task records, source hashes, equal restored initial states,
manual-policy dispositions, provenance, and private-base identities. The JSON
ledger is cross-checked after those raw records; it does not define expected
content. Expected outcomes are bounded prospective predictions, not trials.

## Native reproduction

Requires a disposable Linux/amd64 Tekton v1.6.0 cluster/context, `kubectl`, a local
Docker image cache, PyYAML, and a new isolated namespace. Run the original build
from the pinned parent generator commit or this revision's unchanged Go code:

```sh
go build -o /tmp/retryctl ./cmd/retryctl
python3 experiments/public-oci/live.py --setup
python3 experiments/public-oci/prepare.py
python3 experiments/public-oci/live.py --run --retryctl /tmp/retryctl
python3 experiments/public-oci/capture_final.py
python3 experiments/public-oci/audit.py
```

`prepare.py` assumes the exact pinned host images in
`upstream/runtime-images.json` and the pinned newer builder from
`results/public-oci/builder-image-inspect.json`, plus a bare public Skaffold checkout at
`/tmp/tekton-public-skaffold.git`. Clone that repository at the commit in
`upstream/pins.json`; verify HEAD before proceeding. It exports Linux/amd64
image archives, checks uploaded byte hashes, and serves the public Git snapshot
locally. The test environment has no public pod/controller egress. Local OCI
mirror manifest digests can differ after archive transport; uncompressed layer
identities and effective image configuration are checked against pinned origin
images. Go/Alpine base archives are hundreds of MB and are not duplicated in
this repository; origin digest/config/layer identities and transport checksums
are retained. Reproduction uses a fresh namespace/context, adjusting the fixed
namespace/context in `prepare.py` and `capture_final.py` consistently with CLI
arguments. Never run in a shared/production namespace.

## Evidence map and exclusions

- `upstream/`: pinned original Catalog tasks, Skaffold source, registry API, pins.
- `profile-{tagged,digest}.json`: actual resolved native Pipeline objects.
- `assessment-{1,2,3}.md`: independent automated source-only requirements and
  later input-binding qualifications; not human authors or reviewer ground truth.
- `results/public-oci/{tagged,digest}/source-*`: exact evaluated failed sources,
  child records, build image bytes and logs; captures are unchanged within worlds.
- Per world/arm directories: initial registry, plan/generated run, native run and
  children, delivered manifest/config/layer bytes. `manifests/` and `blobs/` are
  content-addressed bytes; `observations.json` indexes native observations.
- `offline-transports.json`, `base-*`, `end-input-bindings.json`: transport and
  fixed-source/private-base evidence.
- `calibration/`: first source failed before build Pod creation due to controller
  public-registry entrypoint lookup; second compiled the application but old
  Kaniko failed during multistage filesystem cleanup. Neither published the
  source image or ran recovery arms. Both actual native attempts and logs remain
  identifiable, separately from the 16 evaluated arms and two evaluated sources.
- Initial pod mirroring/public-egress and archive-transfer setup errors also
  preceded recovery execution. `protocol.md` records binding refinements.

Source-only predictions were independently locked before native source attempts;
post-calibration execution bindings followed those attempts. This is transparent
prospective source assessment with calibrated transport/runtime, not a formal
preregistration. One execution per cell supports behavior and invocation counts,
not probability, timing/energy improvements, authoring effort or broad safety.

The upstream runtime-image inventory describes initial preparation, including the historical default builder. The evaluated builder binding is explicit in the final profiles and independently checked against `builder-image-inspect.json` and native TaskRun step image IDs. The final namespace inventory accounts for all 20 native PipelineRuns: 16 arms, two evaluated failed sources, and two excluded failed calibrations.
