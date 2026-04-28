import { of } from 'rxjs';
import { HomeComponent } from './home.component';
import { ApiService } from '../../services/api.service';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { Project } from '../../models/project.model';
import { Router } from '@angular/router';

const projects: Project[] = [
  {
    id: 1,
    name: 'Marketing Launch',
    description: 'Campaign planning',
    created_at: '',
    updated_at: '',
    member_count: 1,
  },
  {
    id: 2,
    name: 'Engineering Roadmap',
    description: 'Shared delivery plan',
    created_at: '',
    updated_at: '',
    member_count: 3,
  },
];

describe('HomeComponent board search and filters', () => {
  let component: HomeComponent;
  let apiSpy: jasmine.SpyObj<ApiService>;
  let authSpy: jasmine.SpyObj<AuthService>;
  let routerSpy: jasmine.SpyObj<Router>;

  beforeEach(() => {
    localStorage.clear();

    apiSpy = jasmine.createSpyObj<ApiService>('ApiService', [
      'getProjects',
      'userHasProjectAccess',
      'getProjectMemberCount',
      'seedProjectOwner',
    ]);
    authSpy = jasmine.createSpyObj<AuthService>('AuthService', ['getCurrentUser']);
    routerSpy = jasmine.createSpyObj<Router>('Router', ['navigate']);

    authSpy.getCurrentUser.and.returnValue({ id: '1', name: 'Alice', email: 'alice@example.com' });
    apiSpy.getProjects.and.returnValue(of([]));
    apiSpy.userHasProjectAccess.and.returnValue(true);
    apiSpy.getProjectMemberCount.and.returnValue(1);

    component = new HomeComponent(
      routerSpy,
      apiSpy,
      authSpy,
      {} as ThemeService
    );
    component.projects = projects;
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('filters boards by name and description search', () => {
    component.boardSearchQuery = 'campaign';

    expect(component.visibleProjects.map(project => project.name)).toEqual(['Marketing Launch']);
  });

  it('filters boards by solo and shared membership', () => {
    component.boardFilter = 'solo';
    expect(component.visibleProjects.map(project => project.name)).toEqual(['Marketing Launch']);

    component.boardFilter = 'shared';
    expect(component.visibleProjects.map(project => project.name)).toEqual(['Engineering Roadmap']);
  });

  it('reports and clears active board refinements', () => {
    component.boardSearchQuery = 'roadmap';
    component.boardFilter = 'shared';

    expect(component.hasBoardRefinements).toBeTrue();
    expect(component.boardResultSummary).toBe('Found 1 board matching your view.');

    component.clearBoardRefinements();

    expect(component.boardSearchQuery).toBe('');
    expect(component.boardFilter).toBe('all');
    expect(component.hasBoardRefinements).toBeFalse();
  });
});
