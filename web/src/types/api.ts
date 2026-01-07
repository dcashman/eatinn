// API Response Types based on the EatInn REST API

export interface User {
  id: number;
  created_at: string;
  name: string;
  email: string;
  activated: boolean;
  version: number;
}

export interface Token {
  token: string;
  expiry: string;
}

export interface Recipe {
  id: number;
  created_at: string;
  name: string;
  description: string;
  notes?: string;
  prep_time?: string;  // Duration format from API
  active_time?: string;
  servings?: number;
  ingredients: IngredientEntry[];
  required_equipment: string[];
  instructions: InstructionStep[];
  images: RecipeImage[];
  version: number;
}

export interface IngredientEntry {
  ingredient: string;
  amount?: string;
  unit?: string;
  optional: boolean;
}

export interface InstructionStep {
  step_number: number;
  text: string;
  notes?: string;
}

export interface RecipeImage {
  url: string;
  type: 'thumbnail' | 'main' | 'step';
  step_number?: number;
}

export interface Metadata {
  current_page: number;
  page_size: number;
  first_page: number;
  last_page: number;
  total_records: number;
}

// API Response Envelopes
export interface ApiResponse<T> {
  [key: string]: T;
}

export interface ErrorResponse {
  error: string | Record<string, string>;
}

export interface RecipeListResponse {
  recipes: Recipe[];
  metadata: Metadata;
}

// API Request Types
export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface CreateRecipeRequest {
  name: string;
  description: string;
  notes?: string;
  prep_time?: string;
  active_time?: string;
  servings?: number;
  ingredients: IngredientEntry[];
  required_equipment: string[];
  instructions: InstructionStep[];
  images?: RecipeImage[];
}

export interface RecipeFilters {
  name?: string;
  ingredients?: string[];
  required_equipment?: string[];
  prep_time?: number;
  active_time?: number;
  sort?: string;
  page?: number;
  page_size?: number;
}
