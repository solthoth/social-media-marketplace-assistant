import { NgFor, NgIf } from '@angular/common';
import {
  Component,
  EventEmitter,
  OnDestroy,
  Output,
  inject,
  input,
  signal
} from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import {
  ApiClientService,
  EnrichmentJob,
  InventoryItem
} from '../../core/api-client.service';

const pollIntervalMilliseconds = 1000;
const maxPollAttempts = 30;

@Component({
  selector: 'smm-item-ai-assist',
  standalone: true,
  imports: [MatButtonModule, NgFor, NgIf],
  template: `
    <section class="ai-assist" aria-labelledby="ai-assist-title">
      <div class="ai-assist-header">
        <div>
          <h2 id="ai-assist-title">AI assist</h2>
          <p class="photo-manager-summary">
            Generate missing details from the title and photos.
          </p>
        </div>
        <button
          matButton="filled"
          class="primary-action button-action"
          type="button"
          data-testid="generate-details"
          [disabled]="!canGenerate() || isWorking()"
          (click)="generateDetails()"
        >
          {{ isWorking() ? 'Generating...' : 'Generate details' }}
        </button>
      </div>

      <p *ngIf="!canGenerate()" class="notice">
        Add a title and at least one photo before generating details.
      </p>
      <p *ngIf="job()" class="notice" aria-live="polite">
        AI job status: {{ job()?.status }}
      </p>
      <p *ngIf="errorMessage()" class="notice error" role="alert">
        {{ errorMessage() }}
      </p>
      <p *ngIf="appliedFields().length > 0" class="notice success">
        Filled {{ appliedFields().join(', ') }}.
      </p>

      <dl *ngIf="job()?.status === 'completed'" class="suggestion-list">
        <div *ngFor="let field of suggestionFields()">
          <dt>{{ field.label }}</dt>
          <dd>{{ field.value || 'No suggestion' }}</dd>
        </div>
      </dl>
    </section>
  `
})
export class ItemAiAssistComponent implements OnDestroy {
  readonly itemId = input.required<string>();
  readonly title = input.required<string>();
  readonly photoCount = input.required<number>();
  @Output() readonly itemApplied = new EventEmitter<InventoryItem>();

  private readonly api = inject(ApiClientService);
  private pollTimer: ReturnType<typeof setTimeout> | null = null;

  protected readonly job = signal<EnrichmentJob | null>(null);
  protected readonly isWorking = signal(false);
  protected readonly errorMessage = signal('');
  protected readonly appliedFields = signal<string[]>([]);

  ngOnDestroy(): void {
    this.clearPollTimer();
  }

  protected canGenerate(): boolean {
    return this.title().trim().length > 0 && this.photoCount() > 0;
  }

  protected generateDetails(): void {
    if (!this.canGenerate()) {
      return;
    }
    this.clearPollTimer();
    this.isWorking.set(true);
    this.errorMessage.set('');
    this.appliedFields.set([]);

    this.api.createItemEnrichmentJob(this.itemId()).subscribe({
      next: (job) => {
        this.job.set(job);
        this.pollJob(job.id, 0);
      },
      error: () => {
        this.isWorking.set(false);
        this.errorMessage.set('AI details could not be started.');
      }
    });
  }

  protected suggestionFields(): Array<{ label: string; value: string }> {
    const suggestion = this.job()?.suggestion;
    if (!suggestion) {
      return [];
    }
    return [
      { label: 'Description', value: suggestion.description },
      { label: 'Category', value: suggestion.category },
      { label: 'Size', value: suggestion.size },
      { label: 'Condition', value: suggestion.condition },
      { label: 'Notes', value: suggestion.notes }
    ];
  }

  private pollJob(jobId: string, attempts: number): void {
    this.pollTimer = setTimeout(() => {
      this.api.getItemEnrichmentJob(this.itemId(), jobId).subscribe({
        next: (job) => {
          this.job.set(job);
          if (job.status === 'completed') {
            this.applyJob(job.id);
            return;
          }
          if (job.status === 'failed') {
            this.isWorking.set(false);
            this.errorMessage.set(job.error_message || 'AI details failed.');
            return;
          }
          if (attempts + 1 >= maxPollAttempts) {
            this.isWorking.set(false);
            this.errorMessage.set(
              'AI details are taking longer than expected.'
            );
            return;
          }
          this.pollJob(job.id, attempts + 1);
        },
        error: () => {
          this.isWorking.set(false);
          this.errorMessage.set('AI details could not be loaded.');
        }
      });
    }, pollIntervalMilliseconds);
  }

  private applyJob(jobId: string): void {
    this.api.applyItemEnrichmentJob(this.itemId(), jobId).subscribe({
      next: (response) => {
        this.job.set(response.job);
        this.appliedFields.set(response.applied_fields);
        this.itemApplied.emit(response.item);
        this.isWorking.set(false);
      },
      error: () => {
        this.isWorking.set(false);
        this.errorMessage.set('AI details could not be applied.');
      }
    });
  }

  private clearPollTimer(): void {
    if (this.pollTimer !== null) {
      clearTimeout(this.pollTimer);
      this.pollTimer = null;
    }
  }
}
