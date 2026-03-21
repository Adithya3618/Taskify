import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { ThemeService } from '../../services/theme.service';

@Component({
  selector: 'app-features',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './features.component.html',
  styleUrls: ['./features.component.scss']
})
export class FeaturesComponent {
  constructor(public themeService: ThemeService) {}

  stats = [
    { value: '3',   label: 'Core Modules' },
    { value: '10+', label: 'Live Capabilities' },
    { value: '100%', label: 'Responsive' },
    { value: '<1s', label: 'Load Time' },
  ];

  features = [
    { icon: '⚡', title: 'Instant Boards',      desc: 'Create and name project boards in seconds. Zero setup, zero friction.',                     color: '#0079bf' },
    { icon: '🎯', title: 'Kanban Stages',        desc: 'Custom columns for any workflow — To Do, In Progress, Done and beyond.',                   color: '#4bce97' },
    { icon: '🔒', title: 'Private Workspaces',   desc: 'Every board is scoped to your account. Your work stays yours, always.',                    color: '#818CF8' },
    { icon: '📋', title: 'Task Cards',            desc: 'Rich task cards with titles, descriptions, and inline quick-edit actions.',                color: '#E879F9' },
    { icon: '📡', title: 'Live Chat',             desc: 'Real-time board messaging powered by WebSocket for instant team sync.',                   color: '#22D3EE' },
    { icon: '🛡️', title: 'Secure Auth',           desc: 'JWT sessions with signup, login, and OTP-verified password reset.',                      color: '#f5a623' },
  ];

  roadmap = [
    { icon: '📅', title: 'Due dates & priorities',   status: 'planned' },
    { icon: '🖱️', title: 'Drag-and-drop reorder',    status: 'planned' },
    { icon: '👥', title: 'Team collaboration',        status: 'planned' },
    { icon: '📊', title: 'Analytics dashboard',       status: 'soon'    },
    { icon: '🔔', title: 'Smart notifications',       status: 'soon'    },
    { icon: '🤖', title: 'AI task suggestions',       status: 'future'  },
  ];
}
