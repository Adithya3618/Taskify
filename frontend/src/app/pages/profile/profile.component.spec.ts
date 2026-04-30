import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { RouterTestingModule } from '@angular/router/testing';
import { ProfileComponent } from './profile.component';
import { AuthService, AuthUser } from '../../services/auth.service';

const mockUser: AuthUser = {
  id: '1',
  name: 'Alice Example',
  email: 'alice@example.com',
  role: 'project_admin'
};

describe('ProfileComponent', () => {
  let fixture: ComponentFixture<ProfileComponent>;
  let component: ProfileComponent;
  let authSpy: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    authSpy = jasmine.createSpyObj('AuthService', ['getCurrentUser', 'updateCurrentUser']);
    authSpy.getCurrentUser.and.returnValue(mockUser);

    await TestBed.configureTestingModule({
      imports: [ProfileComponent, RouterTestingModule],
      providers: [{ provide: AuthService, useValue: authSpy }],
    }).compileComponents();

    fixture = TestBed.createComponent(ProfileComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should render the current user details', () => {
    const text = fixture.nativeElement.textContent;

    expect(component.displayName).toBe('Alice Example');
    expect(component.email).toBe('alice@example.com');
    expect(component.role).toBe('Project Admin');
    expect(component.userInitial).toBe('A');
    expect(text).toContain('Alice Example');
    expect(text).toContain('alice@example.com');
  });

  it('should save valid profile changes through AuthService', () => {
    const updatedUser = { ...mockUser, name: 'Alice Updated', email: 'alice.updated@example.com' };
    authSpy.updateCurrentUser.and.returnValue(updatedUser);

    component.profileName = updatedUser.name;
    component.profileEmail = updatedUser.email;
    component.saveProfile();

    expect(authSpy.updateCurrentUser).toHaveBeenCalledOnceWith({
      name: updatedUser.name,
      email: updatedUser.email,
    });
    expect(component.currentUser).toEqual(updatedUser);
    expect(component.profileMessage).toBe('Profile updated.');
    expect(component.profileMessageType).toBe('success');
  });

  it('should reject invalid email before saving', () => {
    component.profileName = 'Alice Updated';
    component.profileEmail = 'not-an-email';
    component.saveProfile();

    expect(authSpy.updateCurrentUser).not.toHaveBeenCalled();
    expect(component.profileMessage).toBe('Enter a valid email address.');
    expect(component.profileMessageType).toBe('error');
  });

  it('should disable save while the form is unchanged', () => {
    const saveButton = fixture.debugElement.query(By.css('.primaryButton')).nativeElement as HTMLButtonElement;

    expect(component.hasProfileChanges).toBeFalse();
    expect(saveButton.disabled).toBeTrue();
  });
});
