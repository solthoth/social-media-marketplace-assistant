import { Component } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatToolbarModule } from '@angular/material/toolbar';
import { RouterLink, RouterOutlet } from '@angular/router';

@Component({
  selector: 'smm-root',
  standalone: true,
  imports: [MatButtonModule, MatToolbarModule, RouterLink, RouterOutlet],
  template: `
    <main class="app-shell">
      <mat-toolbar class="topbar">
        <a class="brand" matButton routerLink="/">Marketplace Assistant</a>
        <nav aria-label="Primary navigation">
          <a matButton routerLink="/">Dashboard</a>
          <a matButton routerLink="/items">Items</a>
          <a matButton="filled" routerLink="/items/new">New item</a>
        </nav>
      </mat-toolbar>

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
