# Independently written policies (automated authors)

Eighteen policies for six Pipelines, three per Pipeline, each written in a separate
`claude -p` session (an AI coding assistant) whose working directory held only the
Pipeline definition, `inputs/policy-semantics.md` and, where relevant, one environment
note (`inputs/*.md`). Sessions had no access to the paper, the project's policies or any
recovery outcome; the prompt is in `run_author.sh`. Fault-injection lines were removed
from the IV-G scripts given to authors. Authors are not practitioners and the environment
notes were written by the project, so this tests whether the contract can be written
correctly from that knowledge, not whether real authors have it.

`score.py` plans every executed failure with each policy and matches the plan's task
actions to arms already executed natively (IV-G closure replication, external-service
experiment, public OCI case). The digest-profile closures match no executed arm and are
not scored. `konflux-*.json` are offline replays of the Konflux template.
