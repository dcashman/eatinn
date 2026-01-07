import { api, ApiError } from '../api/client';
import { router } from '../lib/router';
import { authService } from '../lib/auth';
import type { Recipe, CreateRecipeRequest, IngredientEntry, InstructionStep } from '../types/api';

export function renderRecipeFormPage(params?: Record<string, string>) {
  if (!authService.requireAuth()) return;

  const isEdit = params?.id !== undefined;
  const recipeId = isEdit ? parseInt(params!.id) : undefined;

  const app = document.querySelector<HTMLDivElement>('#app')!;

  app.innerHTML = `
    <div class="recipe-form-container">
      <button id="back-btn" class="btn-secondary">← Back</button>
      <h1>${isEdit ? 'Edit' : 'Create'} Recipe</h1>

      <form id="recipe-form" class="recipe-form">
        <div class="form-section">
          <h2>Basic Information</h2>

          <div class="form-group">
            <label for="name">Recipe Name *</label>
            <input type="text" id="name" name="name" required>
          </div>

          <div class="form-group">
            <label for="description">Description *</label>
            <textarea id="description" name="description" rows="4" required></textarea>
          </div>

          <div class="form-group">
            <label for="notes">Notes</label>
            <textarea id="notes" name="notes" rows="3"></textarea>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="prep_time">Prep Time</label>
              <input type="text" id="prep_time" name="prep_time" placeholder="30m">
              <small>Use format: 30m, 1h30m, 2h</small>
            </div>

            <div class="form-group">
              <label for="active_time">Active Time</label>
              <input type="text" id="active_time" name="active_time" placeholder="45m">
              <small>Use format: 30m, 1h30m, 2h</small>
            </div>

            <div class="form-group">
              <label for="servings">Servings</label>
              <input type="number" id="servings" name="servings" min="1">
            </div>
          </div>
        </div>

        <div class="form-section">
          <h2>Ingredients</h2>
          <div id="ingredients-list"></div>
          <button type="button" id="add-ingredient-btn" class="btn-secondary">+ Add Ingredient</button>
        </div>

        <div class="form-section">
          <h2>Equipment</h2>
          <div id="equipment-list"></div>
          <button type="button" id="add-equipment-btn" class="btn-secondary">+ Add Equipment</button>
        </div>

        <div class="form-section">
          <h2>Instructions</h2>
          <div id="instructions-list"></div>
          <button type="button" id="add-instruction-btn" class="btn-secondary">+ Add Step</button>
        </div>

        <div id="form-error" class="error-message"></div>

        <div class="form-actions">
          <button type="submit" class="btn-primary">${isEdit ? 'Update' : 'Create'} Recipe</button>
          <button type="button" id="cancel-btn" class="btn-secondary">Cancel</button>
        </div>
      </form>
    </div>
  `;

  setupFormHandlers(recipeId);

  const backBtn = app.querySelector('#back-btn')!;
  backBtn.addEventListener('click', () => {
    if (isEdit) {
      router.navigate(`/recipes/${recipeId}`);
    } else {
      router.navigate('/recipes');
    }
  });

  const cancelBtn = app.querySelector('#cancel-btn')!;
  cancelBtn.addEventListener('click', () => {
    if (isEdit) {
      router.navigate(`/recipes/${recipeId}`);
    } else {
      router.navigate('/recipes');
    }
  });

  // Load existing recipe data if editing
  if (isEdit && recipeId) {
    loadRecipeForEdit(recipeId);
  } else {
    // Add initial empty fields
    addIngredientField();
    addEquipmentField();
    addInstructionField();
  }
}

function setupFormHandlers(recipeId?: number) {
  const form = document.querySelector<HTMLFormElement>('#recipe-form')!;
  const addIngredientBtn = document.querySelector('#add-ingredient-btn')!;
  const addEquipmentBtn = document.querySelector('#add-equipment-btn')!;
  const addInstructionBtn = document.querySelector('#add-instruction-btn')!;

  addIngredientBtn.addEventListener('click', () => addIngredientField());
  addEquipmentBtn.addEventListener('click', () => addEquipmentField());
  addInstructionBtn.addEventListener('click', () => addInstructionField());

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    await handleFormSubmit(recipeId);
  });
}

function addIngredientField(ingredient?: IngredientEntry) {
  const list = document.querySelector('#ingredients-list')!;
  const div = document.createElement('div');
  div.className = 'ingredient-item';
  div.innerHTML = `
    <input type="text" placeholder="Ingredient name" class="ingredient-name" value="${ingredient?.ingredient || ''}" required>
    <input type="text" placeholder="Amount" class="ingredient-amount" value="${ingredient?.amount || ''}">
    <input type="text" placeholder="Unit" class="ingredient-unit" value="${ingredient?.unit || ''}">
    <label><input type="checkbox" class="ingredient-optional" ${ingredient?.optional ? 'checked' : ''}> Optional</label>
    <button type="button" class="btn-remove">Remove</button>
  `;
  list.appendChild(div);

  div.querySelector('.btn-remove')!.addEventListener('click', () => div.remove());
}

function addEquipmentField(equipment?: string) {
  const list = document.querySelector('#equipment-list')!;
  const div = document.createElement('div');
  div.className = 'equipment-item';
  div.innerHTML = `
    <input type="text" placeholder="Equipment name" class="equipment-name" value="${equipment || ''}" required>
    <button type="button" class="btn-remove">Remove</button>
  `;
  list.appendChild(div);

  div.querySelector('.btn-remove')!.addEventListener('click', () => div.remove());
}

function addInstructionField(instruction?: InstructionStep) {
  const list = document.querySelector('#instructions-list')!;
  const stepNumber = list.children.length + 1;
  const div = document.createElement('div');
  div.className = 'instruction-item';
  div.innerHTML = `
    <div class="step-number">${stepNumber}</div>
    <div class="step-content">
      <textarea placeholder="Instruction text" class="instruction-text" rows="3" required>${instruction?.text || ''}</textarea>
      <input type="text" placeholder="Notes (optional)" class="instruction-notes" value="${instruction?.notes || ''}">
    </div>
    <button type="button" class="btn-remove">Remove</button>
  `;
  list.appendChild(div);

  div.querySelector('.btn-remove')!.addEventListener('click', () => {
    div.remove();
    renumberInstructions();
  });
}

function renumberInstructions() {
  const list = document.querySelector('#instructions-list')!;
  Array.from(list.children).forEach((item, index) => {
    const stepNum = item.querySelector('.step-number');
    if (stepNum) stepNum.textContent = (index + 1).toString();
  });
}

async function loadRecipeForEdit(id: number) {
  try {
    const recipe = await api.getRecipe(id);

    // Populate basic fields
    (document.querySelector('#name') as HTMLInputElement).value = recipe.name;
    (document.querySelector('#description') as HTMLTextAreaElement).value = recipe.description;
    if (recipe.notes) (document.querySelector('#notes') as HTMLTextAreaElement).value = recipe.notes;
    if (recipe.prep_time) (document.querySelector('#prep_time') as HTMLInputElement).value = recipe.prep_time;
    if (recipe.active_time) (document.querySelector('#active_time') as HTMLInputElement).value = recipe.active_time;
    if (recipe.servings) (document.querySelector('#servings') as HTMLInputElement).value = recipe.servings.toString();

    // Populate ingredients
    recipe.ingredients.forEach(ing => addIngredientField(ing));

    // Populate equipment
    recipe.required_equipment.forEach(eq => addEquipmentField(eq));

    // Populate instructions
    recipe.instructions
      .sort((a, b) => a.step_number - b.step_number)
      .forEach(step => addInstructionField(step));

  } catch (error) {
    console.error('Failed to load recipe:', error);
    alert('Failed to load recipe for editing');
    router.navigate('/recipes');
  }
}

async function handleFormSubmit(recipeId?: number) {
  const errorDiv = document.querySelector<HTMLDivElement>('#form-error')!;
  errorDiv.textContent = '';

  try {
    const formData = collectFormData();

    let recipe: Recipe;
    if (recipeId) {
      recipe = await api.updateRecipe(recipeId, formData);
    } else {
      recipe = await api.createRecipe(formData);
    }

    router.navigate(`/recipes/${recipe.id}`);
  } catch (error) {
    if (error instanceof ApiError) {
      if (typeof error.data.error === 'object') {
        errorDiv.textContent = Object.entries(error.data.error)
          .map(([field, msg]) => `${field}: ${msg}`)
          .join(', ');
      } else {
        errorDiv.textContent = error.message;
      }
    } else {
      errorDiv.textContent = 'An unexpected error occurred';
    }
  }
}

function collectFormData(): CreateRecipeRequest {
  const form = document.querySelector<HTMLFormElement>('#recipe-form')!;

  // Collect ingredients
  const ingredients: IngredientEntry[] = [];
  form.querySelectorAll('.ingredient-item').forEach(item => {
    const ingredient = (item.querySelector('.ingredient-name') as HTMLInputElement).value;
    const amount = (item.querySelector('.ingredient-amount') as HTMLInputElement).value;
    const unit = (item.querySelector('.ingredient-unit') as HTMLInputElement).value;
    const optional = (item.querySelector('.ingredient-optional') as HTMLInputElement).checked;

    if (ingredient) {
      ingredients.push({
        ingredient,
        amount: amount || undefined,
        unit: unit || undefined,
        optional
      });
    }
  });

  // Collect equipment
  const equipment: string[] = [];
  form.querySelectorAll('.equipment-item').forEach(item => {
    const name = (item.querySelector('.equipment-name') as HTMLInputElement).value;
    if (name) equipment.push(name);
  });

  // Collect instructions
  const instructions: InstructionStep[] = [];
  form.querySelectorAll('.instruction-item').forEach((item, index) => {
    const text = (item.querySelector('.instruction-text') as HTMLTextAreaElement).value;
    const notes = (item.querySelector('.instruction-notes') as HTMLInputElement).value;

    if (text) {
      instructions.push({
        step_number: index + 1,
        text,
        notes: notes || undefined
      });
    }
  });

  return {
    name: (form.querySelector('#name') as HTMLInputElement).value,
    description: (form.querySelector('#description') as HTMLTextAreaElement).value,
    notes: (form.querySelector('#notes') as HTMLTextAreaElement).value || undefined,
    prep_time: (form.querySelector('#prep_time') as HTMLInputElement).value || undefined,
    active_time: (form.querySelector('#active_time') as HTMLInputElement).value || undefined,
    servings: parseInt((form.querySelector('#servings') as HTMLInputElement).value) || undefined,
    ingredients,
    required_equipment: equipment,
    instructions
  };
}
