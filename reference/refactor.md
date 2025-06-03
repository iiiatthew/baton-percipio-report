# Refactor Baton Percipio Connector Project

## Summary

The baton-percipio-reportconnector integrates the Percipio LMS system with ConductorOne. It syncs users, courses, and the users' learning activity (course enrollement and completion status) into a ConductorOne app instance. In order to build the C1 resources, the connector currently calls the ApiPathUsersList to get users and the ApiPathCoursesList to get courses. It then makes an API request for a learning report to be generated, then calls the report endpoint with the report ID to fetch the generated report. There are multiple problems with this process, and the relevant user and course information for building those resources can actually be found in the report, so my simplified solution to refactor this connector code to remove calls to the user and courses endpoints and only call the report generation request and report retrieval endpoints, and to use the report data for the userBuilder and courseBuilder functions, instead.

## Rules

- All SDK/C1 interaction functionality is to remain as-is, only change from where the user and course data is being obtained.
- Pare down the user resource to: user.Id, user.Email, user.FistName, and user.LastName (which we will get from the LearnerActivity report (ReportEntry struct) - (we don't need to build the user profile map anymore either)
- Pare down the course resource to: course.ID and courseName (which we will get from the LearnerActivity schema: contentId, and contentTitle)
- we no longer need any client-side pagination, or limitCourses or any code that relates to the user and course api endpoints

## High-Level (Rough) Old Connector Flow (not exhaustive)

1. define resources (user resource, course resource)
2. userBuilder -> getUsers (api call)
3. courseBuilder -> getCourses (api call)
4. define entitlements, get grants, etc.
   - request report api call
   - poll report endpoint until report is available
   - process report and builld report cache
   - define/get/build entitlements and grants from reportCache (users to courses with entitlements like assigned, completed, inProgress)
5. process and build c1z/sync with conductorone via sdk

## New Desired Connector Flow (Rough High-Level)

1. define resources (user resource, course resource)
2. getReport (same as existing process)
   - request report via report request entpoint
   - retreive report via report endpoint
   - build reportCache
3. userBuilder -> get user data from report (ref learnerActivity schema)
4. courseBuilder -> get course data from report (ref learnerActivity schema)
5. define entitlements, get grants, etc. from report
6. process and build c1z/sync with conductorone via sdk

## References

- learnerActivity schema: ./refactor/learner_activity_schema.md
