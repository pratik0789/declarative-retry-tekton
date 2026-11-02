# The failed PipelineRun

Failed task: gate

Succeeded: observe, prepare, build

Did not run: publish

The operator also wants recovery to rerun prepare so that it picks up fresh input; prepare is therefore already on the list.
