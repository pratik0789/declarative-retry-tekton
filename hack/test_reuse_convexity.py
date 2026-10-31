import unittest

from check_reuse_convexity import inspect_reuse


class ReuseConvexityTest(unittest.TestCase):
    def check(self, tasks, actions, final=None):
        pipeline = {"metadata": {"name": "p"}, "spec": {"tasks": tasks, "finally": final or []}}
        plan = {"pipeline": "p", "decision": "recover",
                "tasks": [{"name": n, "action": a} for n, a in actions.items()]}
        return inspect_reuse(pipeline, plan)

    def test_registry_dependency_without_result_or_workspace(self):
        tasks = [{"name": "build"}, {"name": "scan", "runAfter": ["build"]},
                 {"name": "deploy", "runAfter": ["scan"]}]
        warnings = self.check(tasks, {"build": "rerun", "scan": "inherit", "deploy": "rerun"})
        self.assertEqual([(w["task"], w["executedAncestors"], w["executedDescendants"])
                          for w in warnings], [("scan", ["build"], ["deploy"])])
        self.assertEqual(self.check(tasks, {n: "rerun" for n in ["build", "scan", "deploy"]}), [])

    def test_parallel_reuse_is_not_a_gap(self):
        tasks = [{"name": "build"}, {"name": "scan"}, {"name": "deploy", "runAfter": ["build", "scan"]}]
        self.assertEqual(self.check(tasks, {"build": "rerun", "scan": "inherit", "deploy": "continue"}), [])

    def test_embedded_reference_and_final_phase(self):
        tasks = [{"name": "build"}, {"name": "scan", "taskSpec": {
            "steps": [{"script": "check $(tasks.build.results.digest)"}]}}]
        final = [{"name": "report"}]
        warnings = self.check(tasks, {"build": "rerun", "scan": "inherit", "report": "continue"}, final)
        self.assertEqual(warnings[0]["task"], "scan")
        self.assertEqual(warnings[0]["executedDescendants"], ["report"])

    def test_invalid_graph_and_partition_are_rejected(self):
        with self.assertRaisesRegex(ValueError, "unresolved"):
            self.check([{"name": "a", "runAfter": ["missing"]}], {"a": "rerun"})
        with self.assertRaisesRegex(ValueError, "cyclic"):
            self.check([{"name": "a", "runAfter": ["b"]}, {"name": "b", "runAfter": ["a"]}],
                       {"a": "rerun", "b": "inherit"})
        with self.assertRaisesRegex(ValueError, "partition"):
            self.check([{"name": "a"}], {})


if __name__ == "__main__":
    unittest.main()
