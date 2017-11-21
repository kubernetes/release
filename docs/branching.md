# Branching

## Branching from master

Some caveats to branching.

We branch earlier than we really should as a way to encourage stabilization
on the master branch.

During that time only milestone-specific changes are allowed into the
master and when things are stable (enough) we effectively sync the
release branch to the HEAD of the master branch:

*NOTE:* While the new release branch is in this midway state, no alpha
releases should be created off the master branch as this would result in new
master branch tags landing on the release branch.  The tooling in both anago
and in branchff will catch this case, however.  This note is simply an FYI.

```
# branchff <release branch>
$ branchff release-1.8
```

Once that happens we then open up master to milestone+1 changes
and move to the 2 phase [cherry-pick model](https://github.com/kubernetes/community/blob/master/contributors/devel/cherry-picks.md)

## Branching from a tag

For emergency patch releases, we may need to branch from an existing tag
in order to patch in a single change.

Let's take this example.

We need to make a special patch release from the v1.3.0 tag on the release-1.3
branch.

```
# Create the new branch and resulting tags, builds, pushes, notifications...
$ anago release-1.3.0
```

The automatically notified contributor then creates a `cherrypick-candidate` PR
for the emergency change using the `cherry_pick_pull.sh` tool.

```
# Create the new release with the emergency change added only to the v1.3.0 
# tag and branch
$ anago --official release-1.3.0
```
