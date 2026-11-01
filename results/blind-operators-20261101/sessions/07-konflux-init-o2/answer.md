`init` was the first task and it failed, so no task succeeded (failed-run.md says "Succeeded: (none)").
The recovery run will run `init` again because it failed, and every other task runs normally because none of them ran before.
Nothing completed earlier can be reused, so there is nothing to list.

RERUN: []
