# Saved result text must remain data

An independently constructed counterexample uses a successful producer whose saved string is literally `$(tasks.ghost.results.value)` and a failed consumer reading that producer's result. Both the original cited `a2cce22` and previous `26a0ede` supplement emit a reduced Pipeline with only the consumer yet a live reference to nonexistent `ghost`. The original and prior-supplement outputs are retained verbatim, along with exact input bytes.

The correction refuses insertion into retained task definitions whenever a consumed saved string contains `$(`. This deliberately includes parameter/context expressions and escaped-dollar forms; it is a conservative unsupported-input boundary, not a general escaping implementation. `plan` still describes a candidate selection; `create` performs generation checks and fails with exit 1 and empty manifest output for this input. The refusal occurs before creating any Tekton run.

Ordinary saved strings retain existing behavior. Pipeline output values can still use the existing base64 helper, which keeps data out of generated source expressions; the new regression checks its exact saved bytes and reference to the actual publisher. No native execution of this corrected source is claimed. This correction changes no original paper text, numerical observations, cited source identity or historical execution identity.

`verification.json` lists the actual offline checks and source/binary identities. The five syntax regression cases cover missing task, finally, Pipeline param, context and escaped-dollar input. The original scoped suite remains historical; corrected evaluate passes all 45 cases. A refusal is not a successful recovery or evidence of operational value.
