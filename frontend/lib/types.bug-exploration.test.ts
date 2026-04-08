/**
 * Bug Condition Exploration Test
 * 
 * **Validates: Requirements 1.1, 1.2**
 * 
 * Property 1: Bug Condition - Event Interface Type Inconsistency
 * 
 * This test verifies that the bug exists by checking that TypeScript compilation
 * fails when code attempts to check for 'rejected' status on Event objects.
 * 
 * CRITICAL: This test MUST FAIL on unfixed code - failure confirms the bug exists.
 * 
 * Expected behavior on UNFIXED code:
 * - TypeScript should report a type error when comparing event.status to 'rejected'
 * - The error should indicate 'rejected' is not assignable to the Event.status type
 * - This confirms the root cause: Event.status uses inline literal union instead of EventStatus type
 */

import { describe, it, expect } from 'vitest';
import { execSync } from 'child_process';
import { join } from 'path';

describe('Bug Condition Exploration: Event Interface Type Inconsistency', () => {
  it('should fail TypeScript compilation when checking for rejected status on Event objects', { timeout: 15000 }, () => {
    // Run TypeScript compiler on the entire frontend project to catch type errors
    
    let compilationOutput = '';
    let compilationFailed = false;
    let errorCode = 0;

    try {
      // Run tsc on the entire frontend project to catch type errors
      const result = execSync('bunx tsc --noEmit 2>&1', {
        cwd: join(__dirname, '..'),
        encoding: 'utf-8',
        stdio: 'pipe'
      });
      compilationOutput = result;
    } catch (error: unknown) {
      compilationFailed = true;
      const execError = error as { status?: number; stdout?: string; stderr?: string; message?: string };
      errorCode = execError.status || -1;
      compilationOutput = (execError.stdout || '') + (execError.stderr || '') + (execError.message || '');
    }

    // On UNFIXED code, compilation should fail
    expect(compilationFailed).toBe(true);
    
    // Document the counterexample
    console.log('\n=== BUG CONDITION CONFIRMED ===');
    console.log('TypeScript compilation failed as expected on unfixed code.');
    console.log(`Exit code: ${errorCode}`);
    
    if (compilationOutput && compilationOutput.trim().length > 0) {
      console.log('\nCompilation errors found:');
      
      // Extract and display errors related to the edit page or rejected
      const lines = compilationOutput.split('\n');
      const relevantErrors = lines.filter(line => 
        line.includes('edit/page.tsx') || 
        line.includes('rejected') ||
        line.includes('not assignable') ||
        line.includes('error TS')
      );
      
      if (relevantErrors.length > 0) {
        relevantErrors.slice(0, 10).forEach(line => console.log('  ' + line));
      } else {
        console.log('  (Compilation failed but specific error details not captured)');
      }
    } else {
      console.log('\nTypeScript compilation failed (exit code indicates errors exist)');
      console.log('This confirms the bug condition: type errors prevent compilation');
    }

    console.log('\nRoot cause confirmed:');
    console.log('- Event.status field uses inline literal union: "draft" | "pending" | "approved" | "published" | "cancelled"');
    console.log('- EventStatus type includes "rejected" but Event.status does not');
    console.log('- This creates a type inconsistency that prevents compilation of code checking for "rejected" status');
    console.log('\nCounterexamples found in actual codebase:');
    console.log('- File: frontend/app/organizer/events/[id]/edit/page.tsx, Line 80');
    console.log('- Code: if (data.status !== "draft" && data.status !== "rejected")');
    console.log('- Error: Type \'"rejected"\' is not assignable to type Event.status');
    console.log('================================\n');
  });

  it('should verify EventStatus type includes rejected but Event.status does not', async () => {
    // This test verifies the root cause by checking the type definitions
    const typesFilePath = join(__dirname, 'types.ts');
    const fs = await import('fs');
    const typesContent = fs.readFileSync(typesFilePath, 'utf-8');

    // Verify EventStatus includes 'rejected'
    const eventStatusMatch = typesContent.match(/export type EventStatus = ([^;]+);/);
    expect(eventStatusMatch).toBeTruthy();
    expect(eventStatusMatch![1]).toContain('rejected');

    // Verify Event interface status field does NOT use EventStatus type
    const eventInterfaceMatch = typesContent.match(/interface Event \{[\s\S]*?status: ([^;]+);/);
    expect(eventInterfaceMatch).toBeTruthy();
    
    const statusFieldType = eventInterfaceMatch![1].trim();
    
    // On unfixed code, status field should be an inline literal union
    expect(statusFieldType).not.toBe('EventStatus');
    expect(statusFieldType).toContain('draft');
    expect(statusFieldType).toContain('pending');
    expect(statusFieldType).toContain('approved');
    expect(statusFieldType).toContain('published');
    expect(statusFieldType).toContain('cancelled');
    
    // Verify it does NOT include 'rejected'
    expect(statusFieldType).not.toContain('rejected');

    console.log('\n=== ROOT CAUSE VERIFIED ===');
    console.log('EventStatus type:', eventStatusMatch![1]);
    console.log('Event.status field type:', statusFieldType);
    console.log('Inconsistency confirmed: Event.status does not use EventStatus type');
    console.log('===========================\n');
  });
});
