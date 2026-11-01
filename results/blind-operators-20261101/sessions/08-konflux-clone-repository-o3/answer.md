No previously successful task needs to run again. Only `init` succeeded. It writes nothing to any workspace, and in this Pipeline `clone-repository` takes only Pipeline params, not `init`'s results. Copying `init`'s string results into the recovery run is therefore safe.
`clone-repository` runs again anyway because it failed, and every later task never ran, so all of them run normally.
If the workspace is a persistentVolumeClaim, files left by the failed clone persist. `git-clone` normally wipes the folder before cloning (I didn't check its task definition here), so that leftover isn't a reason to rerun anything.
There was no environment note in the directory, so this relies only on the Pipeline file and the operator rules.

RERUN: []
