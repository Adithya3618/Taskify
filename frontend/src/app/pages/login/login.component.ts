import { Component } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss'],
})
export class LoginComponent {
  email = '';
  password = '';
  loading = false;
  error = '';

  constructor(private router: Router) {}

  onSubmit() {
    this.error = '';
    if (!this.email.trim()) {
      this.error = 'Please enter your email.';
      return;
    }
    if (!this.password) {
      this.error = 'Please enter your password.';
      return;
    }
    this.loading = true;
    // TODO: wire to auth API when backend supports it
    setTimeout(() => {
      this.loading = false;
      this.router.navigate(['/']);
    }, 800);
  }
}
