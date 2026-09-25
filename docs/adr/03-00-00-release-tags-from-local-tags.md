# 03-00-00 — release tags bump the highest local tag and survive a failed push

## Context

The repo list gained `v` to tag and push a repo's next major, minor or patch
version. Each repo sets the prefix and suffix its tags wrap around
`MAJOR.MINOR.PATCH` (`v1.2.3`, `mypkg/v1.2.3`, `1.2.3-mypkg`). Those settings,
and the forge links that used to have their own screen on `f`, now share one
per-repo settings form on `3`.

## Decision

**The next version is the highest local tag that exactly matches the repo's
format, bumped.** A tag matches only when it is the prefix, three dot-separated
numbers without leading zeros and the suffix, and nothing else. So `v2.0.0-rc1`
isn't read as a `v`-format release, and several packages in one repo keep
separate version lines. A repo with no matching tag bumps from `0.0.0`. Tags
aren't fetched first, because that would put a network round trip, with its
credential and timeout failures, in front of a screen that only reads.

**The tag is annotated, made at `HEAD`, and pushed alone with
`git push origin refs/tags/<tag>`.** Pushing to `origin` reaches every upstream
through the push URLs lazymux already writes. Branches aren't pushed, so a
release never publishes commits the user didn't mean to push.

**A failed push leaves the local tag in place.** With several upstreams, a push
can land on some forges and fail on others. Deleting the local tag then would
let the next attempt make a different tag object with the same name, which the
forges that already have it would reject. Keeping the tag means the error names
it, `git push origin <tag>` retries it as is, and the next `v` bumps past it.

**Repo settings is a huh form, like the global settings.** `enter` walks the
fields and saves on the last one; `esc` leaves without saving. The old forge
screen saved on `esc`.

## Consequences

- A release someone else pushed isn't seen until `git fetch --tags`. The push
  is then rejected because the tag already exists, and the tag stays local.
- Git must be allowed to create annotated tags, so `user.name` and `user.email`
  have to be set. When they aren't, git's error is shown and nothing is pushed.
- Git is told not to prompt for credentials during the push, as with pull-all,
  so a forge that needs an interactive login fails the push with git's message.
