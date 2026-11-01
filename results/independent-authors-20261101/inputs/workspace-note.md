# Environment

Tasks can fail partway through, after writing some files. The Pipeline’s shared Workspace is bound to a persistentVolumeClaim, so a recovery run sees the files left by the failed run. Operators sometimes explicitly select an already-successful task to rerun so that recovery picks up fresh input.
