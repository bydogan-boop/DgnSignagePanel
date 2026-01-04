export const useApi = () => {
  const fetchWithAuth = async (url, options = {}) => {
    // login.vue'daki isimle aynı: 'auth_token'
    const token = localStorage.getItem('auth_token');
    
    // Header hazırlığı
    const headers = {
      ...options.headers,
      'Content-Type': 'application/json'
    };

    // Eğer token varsa Bearer formatında ekle
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(url, { ...options, headers });

    // 401 hatası gelirse ve şu an login sayfasında değilsek yönlendir
    // Ama önce konsola yaz ki döngüyü görelim
    if (response.status === 401 && !url.includes('/api/login')) {
      console.warn("401 Hatası: Token sunucu tarafından reddedildi!");
      
      // SADECE Dashboard veya veri sayfalarındaysan login'e at
      if (window.location.pathname !== '/login') {
        localStorage.removeItem('auth_token');
        navigateTo('/login');
      }
    }

    return response;
  };

  return { fetchWithAuth };
};