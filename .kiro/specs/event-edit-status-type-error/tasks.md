# Implementation Plan

- [x] 1. Write bug condition exploration test
  - **Property 1: Bug Condition** - Event Interface Type Inconsistency
  - **CRITICAL**: This test MUST FAIL on unfixed code - failure confirms the bug exists
  - **DO NOT attempt to fix the test or the code when it fails**
  - **NOTE**: This test encodes the expected behavior - it will validate the fix when it passes after implementation
  - **GOAL**: Surface TypeScript compilation errors that demonstrate the type inconsistency exists
  - **Scoped PBT Approach**: For this type-level bug, verify TypeScript compilation fails when checking for 'rejected' status on Event objects
  - Test that TypeScript compilation fails when code checks `event.status === 'rejected'` or `event.status !== 'rejected'`
  - Verify the error message indicates 'rejected' is not assignable to the Event.status type
  - Run test on UNFIXED code (current Event interface with inline literal union)
  - **EXPECTED OUTCOME**: Test FAILS with TypeScript error "Type '"rejected"' is not assignable to type..."
  - Document counterexamples found: specific files and line numbers where compilation fails
  - Verify root cause: Event.status field uses inline literal union instead of EventStatus type
  - Mark task complete when test is written, run, and failure is documented
  - _Requirements: 1.1, 1.2_

- [x] 2. Write preservation property tests (BEFORE implementing fix)
  - **Property 2: Preservation** - Existing Status Checks Compile Successfully
  - **IMPORTANT**: Follow observation-first methodology
  - Observe behavior on UNFIXED code for non-buggy inputs (status checks for 'draft', 'pending', 'approved', 'published', 'cancelled')
  - Verify TypeScript compilation succeeds for existing status value checks
  - Document which files successfully compile with current type definition
  - Write test that verifies compilation succeeds for all non-'rejected' status checks
  - Run tests on UNFIXED code
  - **EXPECTED OUTCOME**: Tests PASS (confirms baseline behavior to preserve)
  - Mark task complete when tests are written, run, and passing on unfixed code
  - _Requirements: 3.1, 3.2_

- [x] 3. Fix Event interface status field type inconsistency

  - [x] 3.1 Implement the fix
    - Open `frontend/lib/types.ts`
    - Locate the Event interface (line 66)
    - Change the status field from `status: 'draft' | 'pending' | 'approved' | 'published' | 'cancelled';` to `status: EventStatus;`
    - Verify EventStatus type definition (line 36) includes all required values: 'draft', 'pending', 'approved', 'rejected', 'published', 'cancelled'
    - Save the file
    - _Bug_Condition: isBugCondition(eventInterface) where eventInterface.statusField.type is LiteralUnion AND NOT EventStatus_
    - _Expected_Behavior: Event interface SHALL use EventStatus type for status field, allowing all valid status values including 'rejected'_
    - _Preservation: All existing code checking for 'draft', 'pending', 'approved', 'published', 'cancelled' SHALL continue to compile and function correctly_
    - _Requirements: 1.1, 1.2, 2.1, 2.2, 3.1, 3.2_

  - [x] 3.2 Verify bug condition exploration test now passes
    - **Property 1: Expected Behavior** - Event Interface Uses EventStatus Type
    - **IMPORTANT**: Re-run the SAME test from task 1 - do NOT write a new test
    - The test from task 1 encodes the expected behavior
    - When this test passes, it confirms the expected behavior is satisfied
    - Run TypeScript compilation check for code that checks 'rejected' status
    - Verify compilation succeeds without type errors
    - Verify error message no longer appears
    - **EXPECTED OUTCOME**: Test PASSES (confirms bug is fixed)
    - _Requirements: 2.1, 2.2_

  - [x] 3.3 Verify preservation tests still pass
    - **Property 2: Preservation** - Existing Status Checks Continue Working
    - **IMPORTANT**: Re-run the SAME tests from task 2 - do NOT write new tests
    - Run TypeScript compilation check for all existing status value checks
    - Verify all files that previously compiled still compile successfully
    - Verify no new type errors introduced
    - **EXPECTED OUTCOME**: Tests PASS (confirms no regressions)
    - Confirm all tests still pass after fix (no regressions)

- [x] 4. Checkpoint - Ensure all tests pass
  - Run full TypeScript type check: `bunx tsc --noEmit`
  - Run Next.js production build: `bun run build`
  - Verify no TypeScript compilation errors
  - Verify all status checks compile correctly
  - Ask the user if questions arise
