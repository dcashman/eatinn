import { api, ApiError } from '../api/client';
import { router } from '../lib/router';
import { authService } from '../lib/auth';
import type { Recipe } from '../types/api';

export function renderRecipeDetailPage(params: Record<string, string>) {
  const recipeId = parseInt(params.id);
  const app = document.querySelector<HTMLDivElement>('#app')!;

  app.innerHTML = `
    <div class="recipe-detail-container">
      <button id="back-btn" class="btn-secondary">← Back to Recipes</button>
      <div id="recipe-content">
        <p class="loading">Loading recipe...</p>
      </div>
    </div>
  `;

  const backBtn = app.querySelector('#back-btn')!;
  backBtn.addEventListener('click', () => router.navigate('/recipes'));

  loadRecipeDetail(recipeId);
}

async function loadRecipeDetail(id: number) {
  const container = document.querySelector<HTMLDivElement>('#recipe-content')!;
  const authState = authService.getState();

  try {
    const recipe = await api.getRecipe(id);

    const mainImage = recipe.images?.find(img => img.type === 'main');
    const imageUrl = mainImage?.url || 'https://via.placeholder.com/800x400?text=No+Image';

    container.innerHTML = `
      <div class="recipe-detail">
        <div class="recipe-header">
          <img src="${imageUrl}" alt="${recipe.name}" class="recipe-main-image">
          <div class="recipe-title-section">
            <h1>${recipe.name}</h1>
            ${authState.isAuthenticated ? `
              <div class="recipe-actions">
                <button id="edit-recipe-btn" class="btn-primary">Edit</button>
                <button id="delete-recipe-btn" class="btn-danger">Delete</button>
              </div>
            ` : ''}
          </div>
        </div>

        <div class="recipe-info">
          <p class="recipe-description">${recipe.description}</p>

          <div class="recipe-meta-details">
            ${recipe.prep_time ? `<div><strong>Prep Time:</strong> ${recipe.prep_time}</div>` : ''}
            ${recipe.active_time ? `<strong>Active Time:</strong> ${recipe.active_time}</div>` : ''}
            ${recipe.servings ? `<div><strong>Servings:</strong> ${recipe.servings}</div>` : ''}
          </div>

          ${recipe.notes ? `
            <div class="recipe-notes">
              <h3>Notes</h3>
              <p>${recipe.notes}</p>
            </div>
          ` : ''}
        </div>

        <div class="recipe-sections">
          <div class="ingredients-section">
            <h2>Ingredients</h2>
            <ul class="ingredients-list">
              ${recipe.ingredients.map(ing => `
                <li>
                  ${ing.amount || ''} ${ing.unit || ''} ${ing.ingredient}
                  ${ing.optional ? '<span class="optional-tag">(optional)</span>' : ''}
                </li>
              `).join('')}
            </ul>
          </div>

          ${recipe.required_equipment && recipe.required_equipment.length > 0 ? `
            <div class="equipment-section">
              <h2>Equipment</h2>
              <ul class="equipment-list">
                ${recipe.required_equipment.map(eq => `<li>${eq}</li>`).join('')}
              </ul>
            </div>
          ` : ''}

          <div class="instructions-section">
            <h2>Instructions</h2>
            <ol class="instructions-list">
              ${recipe.instructions.map(step => `
                <li>
                  <div class="step-text">${step.text}</div>
                  ${step.notes ? `<div class="step-notes">${step.notes}</div>` : ''}
                </li>
              `).join('')}
            </ol>
          </div>
        </div>
      </div>
    `;

    // Add event listeners for edit and delete buttons
    const editBtn = container.querySelector('#edit-recipe-btn');
    if (editBtn) {
      editBtn.addEventListener('click', () => router.navigate(`/recipes/${id}/edit`));
    }

    const deleteBtn = container.querySelector('#delete-recipe-btn');
    if (deleteBtn) {
      deleteBtn.addEventListener('click', () => deleteRecipe(id));
    }
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      container.innerHTML = `
        <div class="error-message">
          <h2>Recipe Not Found</h2>
          <p>The recipe you're looking for doesn't exist or has been deleted.</p>
        </div>
      `;
    } else {
      container.innerHTML = `
        <div class="error-message">
          <h2>Error Loading Recipe</h2>
          <p>Failed to load recipe. Please try again.</p>
        </div>
      `;
      console.error('Failed to load recipe:', error);
    }
  }
}

async function deleteRecipe(id: number) {
  if (!confirm('Are you sure you want to delete this recipe? This action cannot be undone.')) {
    return;
  }

  try {
    await api.deleteRecipe(id);
    router.navigate('/recipes');
  } catch (error) {
    alert('Failed to delete recipe. Please try again.');
    console.error('Failed to delete recipe:', error);
  }
}
