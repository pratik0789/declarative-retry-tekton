The only task that ran was `init`, and it failed, so the recovery tool runs it again automatically.
Nothing succeeded, so there is no completed work to reuse and no reused output that could be stale.
Every other task never ran, so they all run normally in the Pipeline's order after `init`.
You don't need to list any tasks.

RERUN: []
