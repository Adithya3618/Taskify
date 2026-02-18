import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';

@Component({
  selector: 'app-error-banner',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './error-banner.component.html',
  styleUrls: ['./error-banner.component.scss']
})
export class ErrorBannerComponent {
  @Input() title: string = 'Something went wrong';
  @Input() message: string = '';
  @Input() hint: string = '';
  @Input() canRetry: boolean = false;

  @Output() retry = new EventEmitter<void>();
  @Output() dismiss = new EventEmitter<void>();

  onRetry() { this.retry.emit(); }
  onDismiss() { this.dismiss.emit(); }
}
