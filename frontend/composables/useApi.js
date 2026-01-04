export const useApi = () => {
  const fetchWithAuth = async (url, options = {}) => {
    const token = localStorage.getItem('auth_token');
    
    // Header'ları kopyala
    const headers = { ...options.headers };

    // EĞER body bir FormData İSE: 
    // Tarayıcının "boundary" değerini kendisinin eklemesi için 
    // Content-Type başlığını tamamen SİLMELİYİZ.
    if (options.body instanceof FormData) {
      delete headers['Content-Type'];
    } else if (!headers['Content-Type']) {
      // Eğer FormData değilse ve manuel atanmamışsa varsayılan JSON yap
      headers['Content-Type'] = 'application/json';
    }

    // Token ekle
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(url, { ...options, headers });

    if (response.status === 401 && !url.includes('/api/login')) {
      localStorage.removeItem('auth_token');
      navigateTo('/login');
    }

    return response;
  };

  return { fetchWithAuth };
};