import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { AuthService, AuthUser } from './auth.service';

const TOKEN_KEY   = 'taskify.auth.token';
const SESSION_KEY = 'taskify.auth.session';

const mockUser: AuthUser = { id: '1', name: 'Alice', email: 'alice@example.com' };
const mockToken = 'mock-jwt-token';

describe('AuthService', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
    });
    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
    localStorage.clear();
  });

  // isAuthenticated()
  it('should return false when no token is stored', () => {
    expect(service.isAuthenticated()).toBeFalse();
  });

  it('should return true when a token exists in localStorage', () => {
    localStorage.setItem(TOKEN_KEY, mockToken);
    expect(service.isAuthenticated()).toBeTrue();
  });

  // getToken()
  it('should return null when no token is set', () => {
    expect(service.getToken()).toBeNull();
  });

  it('should return the token stored in localStorage', () => {
    localStorage.setItem(TOKEN_KEY, mockToken);
    expect(service.getToken()).toBe(mockToken);
  });

  // getCurrentUser()
  it('should return null when no session is stored', () => {
    expect(service.getCurrentUser()).toBeNull();
  });

  it('should return the parsed user from localStorage', () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify(mockUser));
    expect(service.getCurrentUser()).toEqual(mockUser);
  });

  it('should return null when session contains invalid JSON', () => {
    localStorage.setItem(SESSION_KEY, 'not-valid-json{{{');
    expect(service.getCurrentUser()).toBeNull();
  });

  // logout()
  it('should remove token and session from localStorage on logout', () => {
    localStorage.setItem(TOKEN_KEY, mockToken);
    localStorage.setItem(SESSION_KEY, JSON.stringify(mockUser));
    service.logout();
    expect(localStorage.getItem(TOKEN_KEY)).toBeNull();
    expect(localStorage.getItem(SESSION_KEY)).toBeNull();
  });

  // updateCurrentUser()
  it('should return null from updateCurrentUser when no user is logged in', () => {
    expect(service.updateCurrentUser({ name: 'Bob' })).toBeNull();
  });

  it('should merge patch into the current user and persist it', () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify(mockUser));
    const updated = service.updateCurrentUser({ name: 'Alice Updated' });
    expect(updated?.name).toBe('Alice Updated');
    expect(updated?.email).toBe('alice@example.com');
    expect(service.getCurrentUser()?.name).toBe('Alice Updated');
  });

  // login()
  it('should store token and session after successful login', () => {
    service.login('alice@example.com', 'password').subscribe();
    const req = httpMock.expectOne('api/auth/login');
    expect(req.request.method).toBe('POST');
    req.flush({ user: mockUser, token: mockToken });
    expect(service.getToken()).toBe(mockToken);
    expect(service.getCurrentUser()).toEqual(mockUser);
  });

  // register()
  it('should store token and session after successful registration', () => {
    service.register({ name: 'Alice', email: 'alice@example.com', password: 'pass' }).subscribe();
    const req = httpMock.expectOne('api/auth/register');
    expect(req.request.method).toBe('POST');
    req.flush({ user: mockUser, token: mockToken });
    expect(service.getToken()).toBe(mockToken);
    expect(service.getCurrentUser()).toEqual(mockUser);
  });

});
