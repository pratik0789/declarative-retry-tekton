# Participant instructions — draft

These instructions describe a planned study. No participant has completed it.

You will work on recovery decisions for three CI pipeline compositions. We are studying the work and correctness of different recovery workflows; we are not testing your job performance. Your workflow assignment and order vary. Some failures admit safe reuse, some require extra tasks, and some require refusal. A safe full restart can be a valid choice even when it repeats additional work.

For each composition, you receive its Task/Pipeline definitions, failure captures, domain-contract facts and state observations. The same facts and recovery generator are available across workflows. You may use the supplied references, keep notes, reuse earlier work, and write scripts. Record any factual clarification you receive and any automation you use.

Depending on the assigned workflow, you will submit explicit task selections, write a generic persisted rule, or write a declarative policy. You may choose refusal and explain why. Submit your first decision, its rationale and confidence before it is checked. You can correct it after the result and feedback. The study records initial and corrected decisions, actions, relevant work time and the supplied feedback. Initial rule creation and later maintenance count as work.

After several incidents, a genuine source revision will be presented. Decide what to update and submit the next recovery decision. Do not assume a previous rule still applies; use the supplied source and state facts. Do not assume a change always requires more tasks.

Please say when information is insufficient. An unknown or unsupported situation should not be silently treated as safe. You may stop; incomplete decisions remain visible in the study record. This exercise does not authorize changing production systems or contacting external services. The actual executable environments and logging details must be supplied before participant sessions begin.
