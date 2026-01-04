export default defineNuxtConfig({
  ssr: false,
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss'],

  // Route Rules: Proxy ayarları
  routeRules: {
    // API İstekleri: Yerelde çalışan Go Backend'e (8081) yönlendirir
    '/api/**': { 
      proxy: 'http://localhost:8081/api/**' 
    },
    
    // Medya İstekleri: Uzak sunucudaki (sn.dgnconception.fr) resimlere yönlendirir
    '/uploads/**': { 
      proxy: 'http://sn.dgnconception.fr:8081/uploads/**' 
    }
  },

  css: ['~/assets/css/main.css'],

  // Cloud ortamlarında dışarıdan erişimi kolaylaştırmak için runtimeConfig eklenebilir
  runtimeConfig: {
    public: {
      apiBase: '/api'
    }
  },

  compatibilityDate: '2024-04-03'
})