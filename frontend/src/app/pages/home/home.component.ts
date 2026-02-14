import { Component } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api.service';
import { Project } from '../../models/project.model';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent {
  projects: Project[] = [];
  loading = true;
  apiError = false;
  useDemoData = false;

  // Predefined colors for boards
  colors = ['#cdb4db', '#bde0fe', '#ffc8dd', '#ffafcc', '#a2d2ff', '#bde0fe'];

  constructor(
    private router: Router,
    private apiService: ApiService
  ) {
    console.log('HomeComponent initialized');
    this.loadProjects();
    
    // Fallback timeout - show boards even if API fails
    setTimeout(() => {
      if (this.loading) {
        console.log('API timeout - showing boards anyway');
        this.loading = false;
      }
    }, 5000);
  }

  loadProjects() {
    console.log('Loading projects from API...');
    this.apiService.getProjects().subscribe({
      next: (projects) => {
        console.log('Projects loaded:', projects);
        this.projects = projects || [];
        this.loading = false;
      },
      error: (err) => {
        console.error('Failed to load projects:', err);
        console.error('Error details:', err.message, err.status, err.url);
        // Use empty array as fallback
        this.apiError = true;
        this.useDemoData = true;
        // Add demo data so user can see the UI
        this.projects = [
          { id: 1, name: 'Demo Project', description: 'This is demo data', created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
          { id: 2, name: 'Sample Board', description: 'Click to open board', created_at: new Date().toISOString(), updated_at: new Date().toISOString() }
        ];
        this.loading = false;
      },
      complete: () => {
        console.log('Projects subscription complete');
      }
    });
  }

  getProjectColor(index: number): string {
    return this.colors[index % this.colors.length];
  }

  openBoard(projectId: number) {
    this.router.navigate(['/board', projectId]);
  }

  createNewBoard() {
    const name = prompt('Enter board name:');
    if (name) {
      this.apiService.createProject({ name, description: '' }).subscribe({
        next: (project) => {
          this.projects.push(project);
        },
        error: (err) => {
          console.error('Failed to create project:', err);
        }
      });
    }
  }
}