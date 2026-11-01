#!/bin/bash
cd "$1" || exit 1
claude -p "You are the author of the Tekton Pipeline in this directory and know the environment described in the environment notes (if present). Read policy-semantics.md, the Pipeline file, and any environment note. Write a PipelineRetryPolicy for this Pipeline. Goal: whenever one task fails (and possibly an operator also selects some already-successful tasks), the recovery run must still produce a correct result; subject to that, reuse as much completed work as possible. Respect every rule in policy-semantics.md. Reply with a short rationale (at most 5 lines), then the policy as a single fenced yaml block. Do not read files outside this directory." --allowedTools Read --output-format text > answer.md 2> err.txt
echo "exit=$?" >> err.txt
