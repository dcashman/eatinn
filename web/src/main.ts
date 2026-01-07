import './style.css';
import { router } from './lib/router';
import { renderLoginPage } from './pages/login';
import { renderRecipesPage } from './pages/recipes';
import { renderRecipeDetailPage } from './pages/recipe-detail';
import { renderRecipeFormPage } from './pages/recipe-form';

// Set up routes
router.addRoute('/', () => {
  // Redirect to recipes page by default
  router.navigate('/recipes');
});

router.addRoute('/login', () => {
  renderLoginPage();
});

router.addRoute('/recipes', () => {
  renderRecipesPage();
});

router.addRoute('/recipes/new', () => {
  renderRecipeFormPage();
});

router.addRoute('/recipes/:id', (params) => {
  renderRecipeDetailPage(params!);
});

router.addRoute('/recipes/:id/edit', (params) => {
  renderRecipeFormPage(params);
});

// Initialize the app
console.log('EatInn frontend initialized');
