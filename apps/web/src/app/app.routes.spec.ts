import { describe, expect, it } from 'vitest';
import { routes } from './app.routes';

describe('routes', () => {
  it('defines initial application routes', () => {
    expect(routes.map((route) => route.path)).toEqual(['', 'items']);
  });
});
