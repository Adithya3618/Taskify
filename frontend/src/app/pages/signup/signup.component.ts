import { Component } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-signup',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './signup.component.html',
  styleUrls: ['./signup.component.scss'],
})
/** Signup page: name, email, password, confirm; link to login; redirects to home on submit. */
export class SignupComponent {
  name = '';
  email = '';
  password = '';
  confirmPassword = '';
  loading = false;
  error = '';

  constructor(
    private router: Router,
    private authService: AuthService
  ) {}

  onSubmit() {
    this.error = '';
    if (!this.name.trim()) {
      this.error = 'Please enter your name.';
      return;
    }
    if (!this.email.trim()) {
      this.error = 'Please enter your email.';
      return;
    }
    if (!this.password) {
      this.error = 'Please enter a password.';
      return;
    }
    if (this.password.length < 6) {
      this.error = 'Password must be at least 6 characters.';
      return;
    }
    if (this.password !== this.confirmPassword) {
      this.error = 'Passwords do not match.';
      return;
    }
    this.loading = true;

    const result = this.authService.register({
      name: this.name,
      email: this.email,
      password: this.password
    });
    this.loading = false;
    if (!result.ok) {
      this.error = result.error || 'Signup failed.';
      return;
    }

    this.router.navigate(['/boards']);
  }
}
