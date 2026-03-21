import { TestBed } from '@angular/core/testing';
import { ThemeService } from './theme.service';

describe('ThemeService', () => {
  let service: ThemeService;

  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
    TestBed.configureTestingModule({});
    service = TestBed.inject(ThemeService);
  });

  afterEach(() => {
    localStorage.clear();
  });

  // isDark getter
  it('should default to dark mode when no saved preference', () => {
    expect(service.isDark).toBeTrue();
  });

  it('should initialise to light mode when localStorage has "light"', () => {
    localStorage.setItem('taskify-theme', 'light');
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({});
    const freshService = TestBed.inject(ThemeService);
    expect(freshService.isDark).toBeFalse();
  });

  // toggle()
  it('should toggle from dark to light', () => {
    expect(service.isDark).toBeTrue();
    service.toggle();
    expect(service.isDark).toBeFalse();
  });

  it('should toggle from light back to dark', () => {
    service.toggle(); // dark → light
    service.toggle(); // light → dark
    expect(service.isDark).toBeTrue();
  });

  it('should persist "light" in localStorage after toggling to light', () => {
    service.toggle();
    expect(localStorage.getItem('taskify-theme')).toBe('light');
  });

  it('should persist "dark" in localStorage after toggling back to dark', () => {
    service.toggle(); // → light
    service.toggle(); // → dark
    expect(localStorage.getItem('taskify-theme')).toBe('dark');
  });

  // _apply() (tested via observable side-effects on document)
  it('should set data-theme="dark" on the html element in dark mode', () => {
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('should set data-theme="light" on the html element after toggling', () => {
    service.toggle();
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });
});
