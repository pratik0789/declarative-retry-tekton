# Per-failure decisions by blind automated operators

Same model and inputs as `results/independent-authors-20261101` (Claude Code 2.1.310,
claude-opus-5-5), but instead of writing one policy per Pipeline, each of three separate
sessions decides, for one failed run, which completed tasks must rerun (`run_operator.sh`).
23 failure situations (the scored IV-G, external-service and registry cases, and the 16
Konflux template positions), 69 sessions. `score_operators.py` plans each selection with an
empty policy plus `--tasks` and matches it to executed arms. Compares deciding at every
failure with writing the decision once; it does not measure human operators.
