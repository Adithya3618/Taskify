import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { ThemeService } from '../../services/theme.service';

@Component({
  selector: 'app-about',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './about.component.html',
  styleUrls: ['./about.component.scss']
})
export class AboutComponent {
  constructor(public themeService: ThemeService) {}

  stats = [
    { value: '60s',  label: 'To your first board' },
    { value: '100%', label: 'Client-side fallback' },
    { value: '0',    label: 'Required setup steps' },
  ];

  values = [
    { icon: '🎯', title: 'Focus First',       desc: 'We strip away noise so teams can focus on delivery, not tool management.',              color: '#0079bf' },
    { icon: '⚡', title: 'Speed Matters',     desc: 'Every interaction is optimized. Fast load, fast save, fast navigation.',              color: '#22D3EE' },
    { icon: '🔒', title: 'Privacy Built-in',  desc: 'Your boards and tasks are private by default. No accidental leaks.',                  color: '#818CF8' },
    { icon: '🌱', title: 'Always Growing',    desc: 'The roadmap is active and features ship regularly based on real feedback.',           color: '#4bce97' },
  ];

  principles = [
    { icon: '🚀', title: 'Zero friction onboarding', desc: "Sign up, create a board, add a task — you're running in under 60 seconds." },
    { icon: '🎨', title: 'Clarity over cleverness',  desc: 'Every UI decision favors clarity. No hidden menus. No guessing games.' },
    { icon: '🛡️', title: 'Resilient by design',     desc: 'When the backend hiccups, the UI stays graceful. Demo mode always works.' },
    { icon: '♿', title: 'Accessible everywhere',    desc: 'Responsive from 320px to 4K. Keyboard navigable. Screen reader friendly.' },
  ];

  stack = ['Angular 17+', 'TypeScript', 'Go', 'SQLite', 'JWT Auth', 'WebSocket', 'SCSS', 'RxJS'];
}
