import { describe, expect, it } from 'vitest';
import { routes } from './app.routes';

describe('routes', () => {
  it('defines initial application routes', () => {
    expect(routes.map((route) => route.path)).toEqual([
      '',
      'items',
      'items/new',
      'items/:id/edit'
    ]);
  });

  it('matches the dashboard only on the full root path', () => {
    expect(routes[0]).toMatchObject({
      path: '',
      pathMatch: 'full'
    });
  });
});
