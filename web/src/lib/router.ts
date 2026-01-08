type RouteHandler = (params?: Record<string, string>) => void;

interface Route {
  pattern: RegExp;
  handler: RouteHandler;
  paramNames: string[];
}

class Router {
  private routes: Route[] = [];
  private currentPath: string = '';

  constructor() {
    window.addEventListener('hashchange', () => this.handleRoute());
    window.addEventListener('load', () => this.handleRoute());
  }

  addRoute(path: string, handler: RouteHandler) {
    // Convert path pattern to regex and extract param names
    // e.g., '/recipe/:id' becomes /^\/recipe\/([^/]+)$/ with params ['id']
    const paramNames: string[] = [];
    const pattern = new RegExp(
      '^' +
      path.replace(/:[^\s/]+/g, (match) => {
        paramNames.push(match.slice(1));
        return '([^/]+)';
      }) +
      '$'
    );

    this.routes.push({ pattern, handler, paramNames });
  }

  navigate(path: string) {
    window.location.hash = '#' + path;
  }

  private handleRoute() {
    const hash = window.location.hash.slice(1) || '/';
    this.currentPath = hash;

    for (const route of this.routes) {
      const match = hash.match(route.pattern);
      if (match) {
        const params: Record<string, string> = {};
        route.paramNames.forEach((name, index) => {
          params[name] = match[index + 1];
        });
        route.handler(params);
        return;
      }
    }

    // No route matched - show 404 or redirect to home
    console.warn('No route matched for:', hash);
  }

  getCurrentPath(): string {
    return this.currentPath;
  }
}

export const router = new Router();
