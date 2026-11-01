#!/bin/bash
cd "$1" || exit 1
claude -p "You operate the Tekton Pipeline in this directory and know the environment described in the environment note (if present). A PipelineRun has failed; failed-run.md describes it. Read operator-semantics.md, the Pipeline file, any environment note, and failed-run.md. Decide which already-successful tasks must run again so that the recovery run produces a correct result, while reusing as much completed work as is safe. Reply with a short rationale (at most 4 lines), then one line of the form RERUN: [task, task, ...] listing only already-successful tasks (use RERUN: [] for none). Do not read files outside this directory." --allowedTools Read --output-format text > answer.md 2> err.txt
echo "exit=$?" >> err.txt
