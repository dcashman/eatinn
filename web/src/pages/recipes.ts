import { api } from '../api/client';
import { router } from '../lib/router';
import { authService } from '../lib/auth';
import type { Recipe } from '../types/api';

export function renderRecipesPage() {
  const app = document.querySelector<HTMLDivElement>('#app')!;
  const authState = authService.getState();

  app.innerHTML = `
    <div class="recipes-container">
      <header class="app-header">
        <h1>EatInn Recipes</h1>
        <div class="header-actions">
          ${authState.isAuthenticated ? `
            <span class="user-email">${authState.userEmail}</span>
            <button id="create-recipe-btn" class="btn-primary">Create Recipe</button>
            <button id="logout-btn" class="btn-secondary">Logout</button>
          ` : `
            <button id="login-btn" class="btn-primary">Login</button>
          `}
        </div>
      </header>

      <div class="search-bar">
        <input type="text" id="search-input" placeholder="Search recipes...">
        <button id="search-btn" class="btn-secondary">Search</button>
      </div>

      <div id="recipes-list" class="recipes-grid">
        <p class="loading">Loading recipes...</p>
      </div>
    </div>
  `;

  // Event listeners
  const loginBtn = app.querySelector('#login-btn');
  if (loginBtn) {
    loginBtn.addEventListener('click', () => router.navigate('/login'));
  }

  const logoutBtn = app.querySelector('#logout-btn');
  if (logoutBtn) {
    logoutBtn.addEventListener('click', () => {
      authService.logout();
      router.navigate('/login');
    });
  }

  const createBtn = app.querySelector('#create-recipe-btn');
  if (createBtn) {
    createBtn.addEventListener('click', () => router.navigate('/recipes/new'));
  }

  const searchBtn = app.querySelector('#search-btn');
  const searchInput = app.querySelector<HTMLInputElement>('#search-input');

  searchBtn?.addEventListener('click', () => {
    loadRecipes({ name: searchInput?.value });
  });

  searchInput?.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
      loadRecipes({ name: searchInput.value });
    }
  });

  // Load recipes
  loadRecipes();
}

async function loadRecipes(filters?: { name?: string }) {
  const listContainer = document.querySelector<HTMLDivElement>('#recipes-list')!;

  try {
    const response = await api.getRecipes(filters);

    if (response.recipes.length === 0) {
      listContainer.innerHTML = '<p class="no-results">No recipes found</p>';
      return;
    }

    listContainer.innerHTML = response.recipes
      .map((recipe) => renderRecipeCard(recipe))
      .join('');

    // Add click handlers to recipe cards
    listContainer.querySelectorAll('.recipe-card').forEach((card) => {
      card.addEventListener('click', () => {
        const id = card.getAttribute('data-id');
        router.navigate(`/recipes/${id}`);
      });
    });
  } catch (error) {
    listContainer.innerHTML = `
      <p class="error-message">Failed to load recipes. Please try again.</p>
    `;
    console.error('Failed to load recipes:', error);
  }
}

function renderRecipeCard(recipe: Recipe): string {
  const thumbnailImage = recipe.images?.find(img => img.type === 'thumbnail');
  const mainImage = recipe.images?.find(img => img.type === 'main');
  const imageUrl = thumbnailImage?.url || mainImage?.url || 'https://via.placeholder.com/300x200?text=No+Image';

  return `
    <div class="recipe-card" data-id="${recipe.id}">
      <img src="${imageUrl}" alt="${recipe.name}" class="recipe-card-image">
      <div class="recipe-card-content">
        <h3>${recipe.name}</h3>
        <p class="recipe-description">${recipe.description.substring(0, 100)}${recipe.description.length > 100 ? '...' : ''}</p>
        <div class="recipe-meta">
          ${recipe.prep_time ? `<span>⏱️ ${recipe.prep_time}</span>` : ''}
          ${recipe.servings ? `<span>🍽️ ${recipe.servings} servings</span>` : ''}
        </div>
      </div>
    </div>
  `;
}
