# Event Edit Status Type Error Bugfix Design

## Overview

This bugfix addresses a TypeScript compilation error in the Next.js production build caused by a type inconsistency in the Event interface. The Event interface's status field is defined with a literal union type that excludes 'rejected', while the EventStatus type includes it. This causes compilation failures when code checks for 'rejected' status values. The fix involves updating the Event interface to use the EventStatus type for its status field, ensuring type consistency across the codebase.

## Glossary

- **Bug_Condition (C)**: The condition that triggers the bug - when the Event interface status field uses a literal union type instead of the EventStatus type
- **Property (P)**: The desired behavior - the Event interface should use EventStatus type for its status field, allowing all valid status values including 'rejected'
- **Preservation**: Existing code that checks for other status values ('draft', 'pending', 'approved', 'published', 'cancelled') must continue to compile and function correctly
- **Event interface**: The TypeScript interface in `frontend/lib/types.ts` (line 66) that defines the structure of event objects
- **EventStatus type**: The TypeScript type alias in `frontend/lib/types.ts` (line 36) that defines all valid event status values
- **Type inconsistency**: When the Event interface's status field type doesn't match the EventStatus type definition

## Bug Details

### Bug Condition

The bug manifests when the Event interface's status field is defined with a literal union type that doesn't include all values from the EventStatus type. Specifically, the status field is defined as `'draft' | 'pending' | 'approved' | 'published' | 'cancelled'` while the EventStatus type includes 'rejected'. This causes TypeScript compilation errors when code attempts to check for 'rejected' status values.

**Formal Specification:**
```
FUNCTION isBugCondition(eventInterface)
  INPUT: eventInterface of type InterfaceDefinition
  OUTPUT: boolean
  
  RETURN eventInterface.statusField.type IS LiteralUnion
         AND eventInterface.statusField.type NOT_EQUAL_TO EventStatus
         AND EXISTS code WHERE code.checksStatus('rejected')
         AND TypeScriptCompiler.fails(code)
END FUNCTION
```

### Examples

- **Example 1**: In `frontend/app/organizer/events/[id]/edit/page.tsx` (line 80), the code checks `if (data.status !== 'draft' && data.status !== 'rejected')`. TypeScript compilation fails because 'rejected' is not a valid value for the Event interface's status field.
- **Example 2**: Multiple files display badges and icons based on event status including 'rejected' (e.g., `frontend/app/organizer/page.tsx` line 80-82), but the Event interface doesn't recognize 'rejected' as valid.
- **Example 3**: Moderator pages filter and display rejected events (e.g., `frontend/app/moderator/reviews/page.tsx`), but the type system doesn't allow 'rejected' as a valid Event status.
- **Edge case**: Code that only checks for the five currently defined status values ('draft', 'pending', 'approved', 'published', 'cancelled') compiles successfully but creates an incomplete type definition.

## Expected Behavior

### Preservation Requirements

**Unchanged Behaviors:**
- All existing code that checks for 'draft', 'pending', 'approved', 'published', or 'cancelled' status values must continue to compile and function correctly
- Type checking for Event objects must continue to work for all existing valid status values
- Runtime behavior of status comparisons must remain unchanged

**Scope:**
All code that does NOT involve checking for 'rejected' status should be completely unaffected by this fix. This includes:
- Status checks for the five currently recognized values
- Event creation and updates that use existing status values
- Display logic for non-rejected events
- Any other Event interface properties and their usage

## Hypothesized Root Cause

Based on the bug description and code analysis, the root cause is:

1. **Type Definition Duplication**: The Event interface defines its status field with an inline literal union type instead of referencing the EventStatus type, creating two separate definitions that can drift out of sync.

2. **Incomplete Type Migration**: When 'rejected' was added to the EventStatus type (line 36), the Event interface's status field (line 66) was not updated to use the EventStatus type, leaving it with the old literal union.

3. **Lack of Single Source of Truth**: Having two separate type definitions for event status (EventStatus type and Event.status field) violates the DRY principle and creates maintenance issues.

4. **TypeScript Strict Mode**: The production build uses strict TypeScript checking, which correctly identifies that 'rejected' is not a valid value for the Event interface's status field type.

## Correctness Properties

Property 1: Bug Condition - Event Interface Uses EventStatus Type

_For any_ Event interface definition, the status field SHALL use the EventStatus type instead of a literal union type, ensuring that all valid status values including 'rejected' are recognized by the type system and code checking for 'rejected' status compiles without errors.

**Validates: Requirements 2.1, 2.2**

Property 2: Preservation - Existing Status Checks Continue Working

_For any_ code that checks Event status values other than 'rejected' ('draft', 'pending', 'approved', 'published', 'cancelled'), the fixed type definition SHALL produce exactly the same compilation and runtime behavior as before, preserving all existing functionality for status comparisons and type checking.

**Validates: Requirements 3.1, 3.2**

## Fix Implementation

### Changes Required

Assuming our root cause analysis is correct:

**File**: `frontend/lib/types.ts`

**Interface**: `Event` (line 66)

**Specific Changes**:
1. **Update status field type**: Change the status field from the inline literal union type to use the EventStatus type
   - Current: `status: 'draft' | 'pending' | 'approved' | 'published' | 'cancelled';`
   - Fixed: `status: EventStatus;`

2. **Verify EventStatus type definition**: Ensure the EventStatus type (line 36) includes all required values
   - Should include: 'draft', 'pending', 'approved', 'rejected', 'published', 'cancelled'

3. **No other changes required**: The fix is a single-line change that establishes the EventStatus type as the single source of truth for event status values

## Testing Strategy

### Validation Approach

The testing strategy follows a two-phase approach: first, verify that the bug exists by attempting to compile code that checks for 'rejected' status on the unfixed type definition, then verify the fix resolves the compilation error and preserves existing behavior.

### Exploratory Bug Condition Checking

**Goal**: Confirm the TypeScript compilation error exists BEFORE implementing the fix. Verify that the root cause is the type inconsistency between EventStatus and Event.status.

**Test Plan**: Attempt to compile the Next.js application with the current type definitions. Examine the TypeScript error messages to confirm they reference the 'rejected' status check in the event edit page.

**Test Cases**:
1. **Production Build Test**: Run `npm run build` on unfixed code (will fail with TypeScript error)
2. **Type Check Test**: Run `npx tsc --noEmit` to see all type errors (will show error at line 80 of edit page)
3. **IDE Type Checking**: Open `frontend/app/organizer/events/[id]/edit/page.tsx` in an IDE with TypeScript support (will show red squiggly under 'rejected')
4. **EventStatus Type Inspection**: Verify EventStatus includes 'rejected' but Event.status does not (confirms type inconsistency)

**Expected Counterexamples**:
- TypeScript error: "Type '"rejected"' is not assignable to type '"draft" | "pending" | "approved" | "published" | "cancelled"'"
- Compilation fails at line 80 of `frontend/app/organizer/events/[id]/edit/page.tsx`
- Root cause confirmed: Event.status field doesn't use EventStatus type

### Fix Checking

**Goal**: Verify that after changing the Event interface's status field to use EventStatus type, all code that checks for 'rejected' status compiles successfully.

**Pseudocode:**
```
FOR ALL code WHERE code.checksEventStatus('rejected') DO
  result := TypeScriptCompiler.compile(code, fixedEventInterface)
  ASSERT result.success = true
  ASSERT result.errors.length = 0
END FOR
```

### Preservation Checking

**Goal**: Verify that after the fix, all existing code that checks for other status values continues to compile and function correctly.

**Pseudocode:**
```
FOR ALL code WHERE code.checksEventStatus(status) AND status != 'rejected' DO
  originalResult := TypeScriptCompiler.compile(code, originalEventInterface)
  fixedResult := TypeScriptCompiler.compile(code, fixedEventInterface)
  ASSERT originalResult.success = fixedResult.success
  ASSERT originalResult.behavior = fixedResult.behavior
END FOR
```

**Testing Approach**: Property-based testing is not strictly necessary for this type-only fix, but we should verify compilation succeeds for all existing status checks. Manual verification of key files is sufficient.

**Test Plan**: After applying the fix, verify that TypeScript compilation succeeds and examine key files that use event status to ensure no regressions.

**Test Cases**:
1. **Compilation Success**: Run `npm run build` and verify it completes without TypeScript errors
2. **Type Check Success**: Run `npx tsc --noEmit` and verify no type errors
3. **Existing Status Checks**: Verify files that check for 'draft', 'pending', 'approved', 'published', 'cancelled' still compile correctly
4. **IDE Type Checking**: Verify no TypeScript errors appear in IDE for any event status checks

### Unit Tests

- Verify TypeScript compilation succeeds after the fix
- Verify no type errors in files that check for 'rejected' status
- Verify no type errors in files that check for other status values
- Verify Event interface correctly accepts all EventStatus values

### Property-Based Tests

Property-based testing is not applicable for this bugfix because:
- This is a compile-time type error, not a runtime behavior issue
- The fix involves changing a type definition, not executable code
- TypeScript's type checker provides exhaustive verification at compile time

### Integration Tests

- Run full Next.js production build and verify it completes successfully
- Verify all pages that display event status render correctly
- Verify organizer event edit page allows editing of draft and rejected events
- Verify moderator pages correctly display and filter rejected events
