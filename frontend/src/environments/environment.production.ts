// The production web container proxies /api to the backend through Nginx.
// Keeping this relative avoids exposing the API port and works on HTTP or HTTPS.
export const environment = {
  apiBase: '',
  production: true
};
