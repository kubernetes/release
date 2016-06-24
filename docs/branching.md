# Branching

Some caveats to branching.

We branch earlier than we really should as a way to encourage stabilization
on the master branch.

During that time only milestone-specific changes are allowed into the
master and when things are stable (enough) we effectively sync the
release branch to the HEAD of the master branch:

```
$ branchff release-1.3
```

Once that happens we then open up master to milestone+1 changes
and move to the 2 phase [cherry-pick model](https://github.com/kubernetes/kubernetes/blob/master/docs/devel/cherry-picks.md)
