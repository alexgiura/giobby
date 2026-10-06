"""Tests for branch_cleanup: issue references and the per-branch decision."""
import unittest

from branch_cleanup import decide, referenced_issues


def pr(number, state="closed", merged=True, sha="abc", body="Refs #8"):
    return {"number": number, "state": state, "merged": merged, "head_sha": sha, "body": body}


class ReferencedIssuesTest(unittest.TestCase):
    def test_single_reference(self):
        self.assertEqual(referenced_issues("Adds the endpoint.\n\nRefs #8"), {8})

    def test_several_references(self):
        self.assertEqual(referenced_issues("Refs #8, #9\nRefs #12"), {8, 9, 12})

    def test_case_insensitive(self):
        self.assertEqual(referenced_issues("refs #8"), {8})

    def test_other_mentions_do_not_count(self):
        self.assertEqual(referenced_issues("Relates to #3, see #4"), set())

    def test_empty_body(self):
        self.assertEqual(referenced_issues(None), set())


class DecideTest(unittest.TestCase):
    def test_merged_and_issue_closed_is_deleted(self):
        delete, _ = decide("feat/x", "abc", [pr(30)], {8: "closed"})
        self.assertTrue(delete)

    def test_issue_still_open_keeps_the_branch(self):
        delete, reason = decide("feat/x", "abc", [pr(30)], {8: "open"})
        self.assertFalse(delete)
        self.assertIn("#8", reason)

    def test_all_referenced_issues_must_be_closed(self):
        delete, _ = decide("feat/x", "abc", [pr(30, body="Refs #8, #9")], {8: "closed", 9: "open"})
        self.assertFalse(delete)

    def test_references_from_every_merged_pr_count(self):
        prs = [pr(30, sha="old", body="Refs #8"), pr(31, sha="abc", body="Refs #9")]
        self.assertFalse(decide("feat/x", "abc", prs, {8: "closed", 9: "open"})[0])
        self.assertTrue(decide("feat/x", "abc", prs, {8: "closed", 9: "closed"})[0])

    def test_merged_without_references_is_deleted(self):
        self.assertTrue(decide("chore/x", "abc", [pr(30, body="Cleanup")], {})[0])

    def test_never_merged_keeps_the_branch(self):
        self.assertFalse(decide("feat/x", "abc", [pr(30, merged=False)], {8: "closed"})[0])

    def test_no_pr_keeps_the_branch(self):
        self.assertFalse(decide("feat/x", "abc", [], {})[0])

    def test_open_pr_keeps_the_branch(self):
        prs = [pr(30), pr(31, state="open", merged=False, sha="abc")]
        self.assertFalse(decide("feat/x", "abc", prs, {8: "closed"})[0])

    def test_commits_after_the_merge_keep_the_branch(self):
        delete, reason = decide("feat/x", "new", [pr(30, sha="abc")], {8: "closed"})
        self.assertFalse(delete)
        self.assertIn("commit", reason)

    def test_long_lived_branches_are_never_deleted(self):
        for branch in ("dev", "main", "release"):
            self.assertFalse(decide(branch, "abc", [pr(30)], {8: "closed"})[0])


if __name__ == "__main__":
    unittest.main()
