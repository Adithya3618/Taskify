import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

@Component({
  selector: 'app-features',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './features.component.html',
  styleUrls: ['./features.component.scss']
})
export class FeaturesComponent {
  featureGroups = [
    {
      title: 'Board Management',
      items: ['Create new boards', 'Rename boards from menu', 'Delete boards safely', 'Private user-scoped boards']
    },
    {
      title: 'Kanban Workflow',
      items: ['Stage-based columns', 'Task cards under each stage', 'Quick add task/list actions', 'Streamlined board navigation']
    },
    {
      title: 'Reliability & UX',
      items: ['Connected/demo status badges', 'Fallback data mode', 'Responsive layout', 'Guided login and signup flow']
    }
  ];

  roadmap = [
    'Task due dates and priorities',
    'Drag-and-drop stage transitions',
    'Real backend auth and JWT sessions',
    'Team members and board collaboration'
  ];
}
