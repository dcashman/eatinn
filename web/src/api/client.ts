import type {
  User,
  Token,
  Recipe,
  RecipeListResponse,
  RegisterRequest,
  LoginRequest,
  CreateRecipeRequest,
  RecipeFilters,
  ErrorResponse
} from '../types/api';

const API_BASE_URL = 'http://localhost:4000';

class ApiError extends Error {
  constructor(public status: number, public data: ErrorResponse) {
    super(typeof data.error === 'string' ? data.error : 'API Error');
    this.name = 'ApiError';
  }
}

class ApiClient {
  private baseUrl: string;
  private token: string | null = null;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
    // Load token from localStorage if it exists
    this.token = localStorage.getItem('auth_token');
  }

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('auth_token', token);
    } else {
      localStorage.removeItem('auth_token');
    }
  }

  getToken(): string | null {
    return this.token;
  }

  isAuthenticated(): boolean {
    return this.token !== null;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers,
    });

    const data = await response.json();

    if (!response.ok) {
      throw new ApiError(response.status, data as ErrorResponse);
    }

    return data;
  }

  // User endpoints
  async register(userData: RegisterRequest): Promise<User> {
    const response = await this.request<{ user: User }>('/v1/users', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
    return response.user;
  }

  async activateUser(token: string): Promise<User> {
    const response = await this.request<{ user: User }>('/v1/users/activated', {
      method: 'PUT',
      body: JSON.stringify({ token }),
    });
    return response.user;
  }

  async login(credentials: LoginRequest): Promise<Token> {
    const response = await this.request<{ authentication_token: Token }>(
      '/v1/tokens/authentication',
      {
        method: 'POST',
        body: JSON.stringify(credentials),
      }
    );
    const token = response.authentication_token;
    this.setToken(token.token);
    return token;
  }

  logout() {
    this.setToken(null);
  }

  // Recipe endpoints
  async getRecipes(filters?: RecipeFilters): Promise<RecipeListResponse> {
    const params = new URLSearchParams();

    if (filters) {
      if (filters.name) params.append('name', filters.name);
      if (filters.ingredients) params.append('ingredients', filters.ingredients.join(','));
      if (filters.required_equipment) params.append('equipment', filters.required_equipment.join(','));
      if (filters.prep_time) params.append('prep_time', filters.prep_time.toString());
      if (filters.active_time) params.append('active_time', filters.active_time.toString());
      if (filters.sort) params.append('sort', filters.sort);
      if (filters.page) params.append('page', filters.page.toString());
      if (filters.page_size) params.append('page_size', filters.page_size.toString());
    }

    const queryString = params.toString();
    const endpoint = `/v1/recipes${queryString ? `?${queryString}` : ''}`;

    return await this.request<RecipeListResponse>(endpoint);
  }

  async getRecipe(id: number): Promise<Recipe> {
    const response = await this.request<{ recipe: Recipe }>(`/v1/recipes/${id}`);
    return response.recipe;
  }

  async createRecipe(recipeData: CreateRecipeRequest): Promise<Recipe> {
    const response = await this.request<{ recipe: Recipe }>('/v1/recipes', {
      method: 'POST',
      body: JSON.stringify(recipeData),
    });
    return response.recipe;
  }

  async updateRecipe(id: number, recipeData: Partial<CreateRecipeRequest>): Promise<Recipe> {
    const response = await this.request<{ recipe: Recipe }>(`/v1/recipes/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(recipeData),
    });
    return response.recipe;
  }

  async deleteRecipe(id: number): Promise<void> {
    await this.request<void>(`/v1/recipes/${id}`, {
      method: 'DELETE',
    });
  }

  // Health check
  async healthCheck(): Promise<{ status: string; environment: string; version: string }> {
    return await this.request('/v1/healthcheck');
  }
}

// Export a singleton instance
export const api = new ApiClient();
export { ApiError };
