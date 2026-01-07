# EatInn Frontend

A modern, TypeScript-based web frontend for the EatInn recipe management API.

## Tech Stack

- **Vite** - Fast build tool and dev server
- **TypeScript** - Type-safe JavaScript
- **Vanilla JS** - No framework, pure TypeScript
- **CSS3** - Modern, responsive styling

## Features

### Implemented

✅ **User Authentication**
- User registration with email verification
- Login with JWT token authentication
- Persistent authentication (localStorage)
- Automatic token management

✅ **Recipe Browsing**
- Browse all recipes (public and authenticated views)
- Search recipes by name
- View recipe details with full information
- Responsive recipe cards with images

✅ **Recipe Management (Authenticated Users)**
- Create new recipes
- Edit existing recipes
- Delete recipes
- Dynamic form fields for:
  - Ingredients (with quantity, unit, optional flag)
  - Equipment
  - Step-by-step instructions

✅ **Routing**
- Client-side routing with hash-based navigation
- Routes:
  - `/` - Home (redirects to recipes)
  - `/login` - Login/Register page
  - `/recipes` - Recipe list
  - `/recipes/:id` - Recipe detail
  - `/recipes/new` - Create recipe
  - `/recipes/:id/edit` - Edit recipe

✅ **Responsive Design**
- Mobile-friendly layout
- Adaptive grids and forms
- Touch-friendly buttons

## Project Structure

```
/web
  /src
    /api
      client.ts           # API client with fetch wrapper
      types.ts            # TypeScript interfaces for API
    /lib
      auth.ts             # Authentication state management
      router.ts           # Client-side routing
    /pages
      login.ts            # Login/registration page
      recipes.ts          # Recipe list page
      recipe-detail.ts    # Single recipe view
      recipe-form.ts      # Create/edit recipe form
    main.ts               # Application entry point
    style.css             # Application styles
  index.html              # HTML template
  package.json            # Dependencies
  tsconfig.json           # TypeScript configuration
  vite.config.ts          # Vite configuration (auto-generated)
```

## Getting Started

### Prerequisites

- Node.js (v18 or higher)
- EatInn API server running on `http://localhost:4000`

### Installation

```bash
cd web
npm install
```

### Development

Start the development server with hot reload:

```bash
npm run dev
```

The app will be available at `http://localhost:5173`

### Build for Production

```bash
npm run build
```

The built files will be in the `dist/` directory.

### Preview Production Build

```bash
npm run preview
```

## API Configuration

The API base URL is configured in `/src/api/client.ts`:

```typescript
const API_BASE_URL = 'http://localhost:4000';
```

To change this for production, update the constant or use an environment variable.

## Usage

### First Time Setup

1. Start the EatInn API server (must be running on port 4000)
2. Start the frontend dev server: `npm run dev`
3. Open `http://localhost:5173` in your browser
4. Register a new account
5. Check your email (or server logs) for the activation token
6. Use the activation token to activate your account (via API currently)
7. Log in with your credentials

### Creating a Recipe

1. Log in to your account
2. Click "Create Recipe" button
3. Fill in recipe details:
   - Name and description (required)
   - Prep time, active time, servings
   - Add ingredients with quantities
   - Add required equipment
   - Add step-by-step instructions
4. Click "Create Recipe"

### Browsing Recipes

- All users can browse recipes
- Search by name in the search bar
- Click on a recipe card to view details

### Editing/Deleting Recipes

- Only authenticated users can edit/delete
- Click "Edit" on a recipe detail page
- Modify fields and click "Update Recipe"
- Click "Delete" to remove (with confirmation)

## TypeScript Types

All API responses and requests are fully typed. See `/src/types/api.ts` for:

- `User` - User account data
- `Recipe` - Full recipe structure
- `Token` - Authentication token
- `IngredientEntry` - Ingredient with quantity/unit
- `InstructionStep` - Numbered instruction step
- Request/response types for all API endpoints

## API Error Handling

The frontend handles API errors gracefully:

- Network errors show user-friendly messages
- Validation errors display field-specific feedback
- Authentication errors redirect to login
- 404 errors show "not found" messages

## Security

- Authentication tokens stored in localStorage
- Bearer token sent with all authenticated requests
- Automatic redirect to login if not authenticated
- Token cleared on logout

## Browser Support

- Modern browsers with ES6+ support
- Chrome, Firefox, Safari, Edge (latest versions)

## Development Notes

### No Framework

This project intentionally uses vanilla TypeScript without a framework like React or Vue. This provides:
- Smaller bundle size
- Direct DOM manipulation
- Educational value
- Easy to understand codebase

### State Management

- Authentication state managed via `authService`
- Listeners pattern for auth state changes
- LocalStorage for persistence
- No complex state management library needed

### Routing

- Hash-based routing (`#/recipes`)
- Pattern matching with params (`:id`)
- Simple and effective for SPAs
- No server configuration needed

## Future Enhancements

Potential improvements:

- [ ] Image upload for recipes
- [ ] Advanced filtering (by ingredients, time, etc.)
- [ ] Recipe ratings and comments
- [ ] Shopping list generation
- [ ] Print-friendly recipe view
- [ ] Dark mode support
- [ ] Offline support (PWA)
- [ ] Share recipes feature

## Troubleshooting

**API Connection Issues:**
- Ensure backend is running on port 4000
- Check browser console for CORS errors
- Verify API_BASE_URL in client.ts

**Authentication Issues:**
- Clear localStorage and try logging in again
- Check that activation token was used
- Verify backend authentication is working

**Build Errors:**
- Delete node_modules and package-lock.json
- Run `npm install` again
- Ensure TypeScript version compatibility

## Contributing

This is a learning project following "Let's Go Further" patterns. Feel free to experiment and extend!
