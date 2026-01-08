import { api } from '../api/client';
import type { LoginRequest, RegisterRequest } from '../types/api';

export interface AuthState {
  isAuthenticated: boolean;
  userEmail: string | null;
}

class AuthService {
  private listeners: Array<(state: AuthState) => void> = [];

  getState(): AuthState {
    return {
      isAuthenticated: api.isAuthenticated(),
      userEmail: localStorage.getItem('user_email'),
    };
  }

  subscribe(listener: (state: AuthState) => void) {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter(l => l !== listener);
    };
  }

  private notify() {
    const state = this.getState();
    this.listeners.forEach(listener => listener(state));
  }

  async login(credentials: LoginRequest): Promise<void> {
    await api.login(credentials);
    localStorage.setItem('user_email', credentials.email);
    this.notify();
  }

  async register(userData: RegisterRequest): Promise<void> {
    await api.register(userData);
    // User needs to activate account before they can log in
  }

  logout(): void {
    api.logout();
    localStorage.removeItem('user_email');
    this.notify();
  }

  requireAuth(): boolean {
    if (!this.getState().isAuthenticated) {
      window.location.hash = '#/login';
      return false;
    }
    return true;
  }
}

export const authService = new AuthService();
