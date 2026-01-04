export const useApi = () => {
    const fetchWithAuth = async (url, options = {}) => {
      const token = localStorage.getItem('auth_token');
      
      options.headers = {
        ...options.headers,
        'Authorization': token || ''
      };
  
      const response = await fetch(url, options);
  
      if (response.status === 401 && !url.includes('/api/login')) {
        localStorage.removeItem('auth_token');
        navigateTo('/login');
      }
  
      return response;
    };
  
    return { fetchWithAuth };
  };