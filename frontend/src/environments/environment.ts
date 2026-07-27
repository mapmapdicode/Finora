const getApiBase = (): string => {
  if (typeof window !== 'undefined' && window.localStorage) {
    const custom = window.localStorage.getItem('wealthos.apiBase');
    if (custom) return custom;
  }
  if (typeof window !== 'undefined' && window.location.port === '4205') {
    return 'http://localhost:8085';
  }
  return 'http://localhost:8080';
};

export const environment = {
  get apiBase() {
    return getApiBase();
  },
  production: false
};

