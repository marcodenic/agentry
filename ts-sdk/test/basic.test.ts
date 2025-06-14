import { describe, it, expect } from 'vitest';
import { invoke } from '../src/index';

describe('invoke', () => {
  it('exports function', () => {
    expect(typeof invoke).toBe('function');
  });
});

