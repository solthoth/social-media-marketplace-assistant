import { Component } from '@angular/core';

@Component({
  selector: 'smm-root',
  standalone: true,
  template: `
    <main class="app-shell">
      <section class="intro">
        <p class="eyebrow">Inventory first</p>
        <h1>Marketplace Assistant</h1>
        <p class="summary">
          Capture items once, keep inventory organized, and prepare listings for connected sales channels.
        </p>
      </section>

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

