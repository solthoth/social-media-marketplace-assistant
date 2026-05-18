import { Component } from '@angular/core';

@Component({
  selector: 'smm-dashboard',
  standalone: true,
  template: `
    <section class="page-section">
      <p class="eyebrow">Inventory first</p>
      <h1>Marketplace Assistant</h1>
      <p class="summary">
        Capture items once, keep inventory organized, and prepare listings for
        connected sales channels.
      </p>
    </section>
  `
})
export class DashboardComponent {}
