# A push is legitimate only if $sha is the merge commit of a pull request merged into $branch
# (holds for merge, squash and rebase merges: GitHub sets merge_commit_sha to the commit on the base).
[.[] | select(.merged_at != null and .base.ref == $branch and .merge_commit_sha == $sha)] | length
