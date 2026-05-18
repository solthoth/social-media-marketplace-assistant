import { Component } from '@angular/core';
import { RouterLink, RouterOutlet } from '@angular/router';

@Component({
  selector: 'smm-root',
  standalone: true,
  imports: [RouterLink, RouterOutlet],
  template: `
    <main class="app-shell">
      <header class="topbar">
        <a class="brand" routerLink="/">Marketplace Assistant</a>
        <nav aria-label="Primary navigation">
          <a routerLink="/">Dashboard</a>
          <a routerLink="/items">Items</a>
          <a routerLink="/items/new">New item</a>
        </nav>
      </header>

      <router-outlet />

      <section class="status-panel" aria-label="Application status">
        <div>
          <span class="label">Frontend</span>
          <strong>Ready</strong>
        </div>
        <div>
          <span class="label">Backend</span>
          <strong>Health endpoint: /healthz</strong>
        </div>
      </section>
    </main>
  `
})
export class AppComponent {}
