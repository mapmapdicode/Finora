const getApiBase = (): string => {
  if (typeof window !== 'undefined' && window.localStorage) {
    const custom = window.localStorage.getItem('wealthos.apiBase');
    if (custom) return custom;
  }
  if (typeof window !== 'undefined' && window.location.port === '4205') {
    return 'http://localhost:8085';
  }
  if (typeof window !== 'undefined' && window.location.hostname && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
    return `http://${window.location.hostname}:8080`;
  }
  return 'http://localhost:8080';
};

export const environment = {
  get apiBase() {
    return getApiBase();
  },
  production: false
};
