import { authService } from '../lib/auth';
import { router } from '../lib/router';
import { ApiError } from '../api/client';

export function renderLoginPage() {
  const app = document.querySelector<HTMLDivElement>('#app')!;

  app.innerHTML = `
    <div class="auth-container">
      <h1>EatInn</h1>

      <div class="tabs">
        <button class="tab active" data-tab="login">Login</button>
        <button class="tab" data-tab="register">Register</button>
      </div>

      <!-- Login Form -->
      <form id="login-form" class="auth-form">
        <h2>Login to Your Account</h2>
        <div class="form-group">
          <label for="login-email">Email</label>
          <input type="email" id="login-email" name="email" required>
        </div>
        <div class="form-group">
          <label for="login-password">Password</label>
          <input type="password" id="login-password" name="password" required>
        </div>
        <div id="login-error" class="error-message"></div>
        <button type="submit" class="btn-primary">Login</button>
      </form>

      <!-- Register Form -->
      <form id="register-form" class="auth-form hidden">
        <h2>Create Account</h2>
        <div class="form-group">
          <label for="register-name">Name</label>
          <input type="text" id="register-name" name="name" required>
        </div>
        <div class="form-group">
          <label for="register-email">Email</label>
          <input type="email" id="register-email" name="email" required>
        </div>
        <div class="form-group">
          <label for="register-password">Password</label>
          <input type="password" id="register-password" name="password" required minlength="8">
          <small>Minimum 8 characters</small>
        </div>
        <div id="register-error" class="error-message"></div>
        <div id="register-success" class="success-message"></div>
        <button type="submit" class="btn-primary">Register</button>
      </form>
    </div>
  `;

  // Tab switching
  const tabs = app.querySelectorAll('.tab');
  const loginForm = app.querySelector<HTMLFormElement>('#login-form')!;
  const registerForm = app.querySelector<HTMLFormElement>('#register-form')!;

  tabs.forEach(tab => {
    tab.addEventListener('click', (e) => {
      const target = e.target as HTMLElement;
      const tabName = target.dataset.tab;

      tabs.forEach(t => t.classList.remove('active'));
      target.classList.add('active');

      if (tabName === 'login') {
        loginForm.classList.remove('hidden');
        registerForm.classList.add('hidden');
      } else {
        loginForm.classList.add('hidden');
        registerForm.classList.remove('hidden');
      }
    });
  });

  // Login form handler
  loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const errorDiv = app.querySelector<HTMLDivElement>('#login-error')!;
    errorDiv.textContent = '';

    const formData = new FormData(loginForm);
    const credentials = {
      email: formData.get('email') as string,
      password: formData.get('password') as string,
    };

    try {
      await authService.login(credentials);
      router.navigate('/recipes');
    } catch (error) {
      if (error instanceof ApiError) {
        errorDiv.textContent = error.message;
      } else {
        errorDiv.textContent = 'An unexpected error occurred';
      }
    }
  });

  // Register form handler
  registerForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const errorDiv = app.querySelector<HTMLDivElement>('#register-error')!;
    const successDiv = app.querySelector<HTMLDivElement>('#register-success')!;
    errorDiv.textContent = '';
    successDiv.textContent = '';

    const formData = new FormData(registerForm);
    const userData = {
      name: formData.get('name') as string,
      email: formData.get('email') as string,
      password: formData.get('password') as string,
    };

    try {
      await authService.register(userData);
      successDiv.textContent = 'Registration successful! Check your email to activate your account.';
      registerForm.reset();
    } catch (error) {
      if (error instanceof ApiError) {
        if (typeof error.data.error === 'object') {
          errorDiv.textContent = Object.values(error.data.error).join(', ');
        } else {
          errorDiv.textContent = error.message;
        }
      } else {
        errorDiv.textContent = 'An unexpected error occurred';
      }
    }
  });
}
