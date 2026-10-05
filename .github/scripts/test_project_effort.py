"""Tests for project_effort: working hours and the per-card decisions."""
import unittest
from datetime import datetime, timezone

from project_effort import decide, working_hours


def utc(text):
    return datetime.fromisoformat(text).replace(tzinfo=timezone.utc)


class WorkingHoursTest(unittest.TestCase):
    # Europe/Rome is UTC+2 in October (CEST): 9:00 Rome = 07:00 UTC.

    def test_inside_one_working_day(self):
        self.assertEqual(working_hours(utc("2026-10-06T07:00"), utc("2026-10-06T10:30")), 3.5)

    def test_ignores_time_outside_9_to_18(self):
        # 06:00-20:00 Rome on a Tuesday counts only 9-18.
        self.assertEqual(working_hours(utc("2026-10-06T04:00"), utc("2026-10-06T18:00")), 9.0)

    def test_skips_nights_and_weekends(self):
        # Friday 17:00 Rome -> Monday 10:00 Rome: 1h Friday + 1h Monday.
        self.assertEqual(working_hours(utc("2026-10-09T15:00"), utc("2026-10-12T08:00")), 2.0)

    def test_full_week(self):
        self.assertEqual(working_hours(utc("2026-10-05T07:00"), utc("2026-10-09T16:00")), 45.0)

    def test_end_before_start_is_zero(self):
        self.assertEqual(working_hours(utc("2026-10-06T10:00"), utc("2026-10-06T08:00")), 0.0)


class DecideTest(unittest.TestCase):
    NOW = "2026-10-06T10:00:00Z"

    def test_records_the_start_when_work_begins(self):
        self.assertEqual(decide("In progress", None, None, None, self.NOW), {"start": self.NOW})

    def test_keeps_an_existing_start(self):
        self.assertEqual(decide("In progress", "2026-10-05T07:00:00Z", None, None, self.NOW), {})

    def test_records_the_end_when_the_card_reaches_test(self):
        self.assertEqual(decide("Test", "2026-10-06T07:00:00Z", None, None, self.NOW), {"end": self.NOW})

    def test_work_resumed_after_a_failed_test_clears_the_end(self):
        self.assertEqual(decide("In progress", "2026-10-06T07:00:00Z", "2026-10-06T09:00:00Z", None, self.NOW),
                         {"end": None})

    def test_done_writes_the_effective_hours(self):
        self.assertEqual(decide("Done", "2026-10-06T07:00:00Z", "2026-10-06T10:00:00Z", None, self.NOW),
                         {"effective": 3.0})

    def test_done_without_a_recorded_end_uses_now(self):
        self.assertEqual(decide("Done", "2026-10-06T07:00:00Z", None, None, self.NOW),
                         {"end": self.NOW, "effective": 3.0})

    def test_done_never_overwrites_an_existing_value(self):
        self.assertEqual(decide("Done", "2026-10-06T07:00:00Z", "2026-10-06T10:00:00Z", 2.5, self.NOW), {})

    def test_done_without_a_start_is_left_alone(self):
        self.assertEqual(decide("Done", None, None, None, self.NOW), {})

    def test_backlog_and_todo_are_ignored(self):
        self.assertEqual(decide("Todo", None, None, None, self.NOW), {})
        self.assertEqual(decide("Backlog", None, None, None, self.NOW), {})


if __name__ == "__main__":
    unittest.main()
