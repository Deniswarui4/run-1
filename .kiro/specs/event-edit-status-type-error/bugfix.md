# Bugfix Requirements Document

## Introduction

The Next.js production build fails with a TypeScript compilation error in the organizer event edit page. The error occurs because the code checks for a 'rejected' status value, but the Event interface's status field type does not include 'rejected' as a valid value. This inconsistency between the EventStatus type definition (which includes 'rejected') and the Event interface's status field (which doesn't) prevents the application from building.

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN the Event interface status field is defined as `'draft' | 'pending' | 'approved' | 'published' | 'cancelled'` (line 66 in frontend/lib/types.ts) THEN the system fails to compile code that checks for 'rejected' status

1.2 WHEN the EventStatus type includes 'rejected' but the Event interface does not use this type THEN the system creates a type inconsistency that causes compilation errors

### Expected Behavior (Correct)

2.1 WHEN the Event interface status field is defined THEN the system SHALL use the EventStatus type to ensure consistency

2.2 WHEN code checks for 'rejected' status on an Event object THEN the system SHALL compile without type errors

### Unchanged Behavior (Regression Prevention)

3.1 WHEN code checks for other valid status values ('draft', 'pending', 'approved', 'published', 'cancelled') THEN the system SHALL CONTINUE TO compile and function correctly

3.2 WHEN the EventStatus type is used elsewhere in the codebase THEN the system SHALL CONTINUE TO work with all existing status values
