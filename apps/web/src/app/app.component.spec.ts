import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { describe, expect, it } from 'vitest';
import { AppComponent } from './app.component';
import { routes } from './app.routes';

describe('AppComponent', () => {
  it('creates the application shell', async () => {
    await TestBed.configureTestingModule({
      imports: [AppComponent],
      providers: [provideRouter(routes)]
    }).compileComponents();

    const fixture = TestBed.createComponent(AppComponent);
    fixture.detectChanges();

    expect(fixture.componentInstance).toBeTruthy();
    expect(fixture.nativeElement.querySelector('mat-toolbar')).toBeTruthy();
  });

  it('shows primary navigation for creating inventory items', async () => {
    await TestBed.configureTestingModule({
      imports: [AppComponent],
      providers: [provideRouter(routes)]
    }).compileComponents();

    const fixture = TestBed.createComponent(AppComponent);
    fixture.detectChanges();

    const links = Array.from<HTMLAnchorElement>(
      fixture.nativeElement.querySelectorAll('nav a')
    ).map((link) => ({
      text: link.textContent.trim(),
      href: link.getAttribute('href')
    }));

    expect(links).toContainEqual({ text: 'New item', href: '/items/new' });
  });
});
