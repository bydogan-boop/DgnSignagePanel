<template>
  <div v-if="loading" class="flex flex-col items-center justify-center min-h-screen text-gray-400 bg-[#0f111a]">
    <div class="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-purple-500 mb-4"></div>
    <p class="text-xl">Yükleniyor...</p>
  </div>

  <div v-else-if="error" class="flex flex-col items-center justify-center min-h-screen text-red-400 bg-[#0f111a] p-6 text-center">
    <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
    </svg>
    <p class="text-xl font-bold mb-2">Hata Oluştu</p>
    <p class="mb-6">{{ error }}</p>
    <router-link to="/restaurants" class="bg-purple-600 hover:bg-purple-700 text-white px-6 py-2 rounded-lg transition-colors">
      Restoranlara Geri Dön
    </router-link>
  </div>

  <div v-else-if="restaurant && restaurant.restaurant" class="container mx-auto p-4 md:p-6 min-h-screen">
    
    <div class="mb-6">
      <router-link to="/restaurants" class="inline-flex items-center text-purple-400 hover:text-purple-300 transition-colors group">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 transform group-hover:-translate-x-1 transition-transform" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M9.707 14.707a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 1.414L7.414 9H15a1 1 0 110 2H7.414l2.293 2.293a1 1 0 010 1.414z" clip-rule="evenodd" />
        </svg>
        Tüm Restoranlara Geri Dön
      </router-link>
    </div>

    <div class="bg-gray-800 rounded-2xl shadow-xl p-6 mb-8 border border-gray-700 flex flex-col md:flex-row items-center md:items-start text-center md:text-left">
      <img 
        :src="restaurant.restaurant.logo_url || '/placeholder-logo.svg'" 
        alt="Restoran Logosu" 
        class="w-32 h-32 rounded-full border-4 border-gray-700 object-cover mb-4 md:mb-0 md:mr-8 shadow-2xl"
      >
      <div class="flex-grow">
        <h1 class="text-4xl font-extrabold text-white mb-2">{{ restaurant.restaurant.name }}</h1>
        <p v-if="restaurant.restaurant.phone" class="text-lg text-gray-400 flex items-center justify-center md:justify-start gap-2">
           <span class="text-purple-400">📞</span> {{ restaurant.restaurant.phone }}
        </p>
        <p v-if="restaurant.restaurant.address" class="text-lg text-gray-400 mt-1">
           <span class="text-purple-400">📍</span> {{ restaurant.restaurant.address }}
        </p>
        <p class="text-xs text-gray-500 mt-4 italic">
          Kayıt Tarihi: {{ restaurant.restaurant.created_at ? new Date(restaurant.restaurant.created_at).toLocaleDateString('tr-TR') : '-' }}
        </p>
      </div>
    </div>

    <div class="mt-12">
      <h2 class="text-2xl font-bold text-white mb-6 flex items-center gap-3">
        <span class="bg-purple-600 w-2 h-8 rounded-full"></span>
        Bu Restorana Ait Ekranlar
      </h2>
      
      <div v-if="!restaurant.screens || restaurant.screens.length === 0" class="text-center py-16 text-gray-500 bg-gray-800/50 rounded-2xl border-2 border-dashed border-gray-700">
        <p class="text-lg">Bu restorana ait henüz aktif bir ekran bulunamadı.</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
        <div v-for="screen in restaurant.screens" :key="screen.id" class="bg-gray-800 rounded-2xl shadow-2xl overflow-hidden border border-gray-700 hover:border-purple-500/50 transition-all group">
          <div class="relative aspect-video bg-black">
            <img v-if="screen.content_type === 'image'" :src="screen.media_url" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500">
            <video v-else-if="screen.content_type === 'video'" :src="screen.media_url" class="w-full h-full object-cover" muted autoplay loop></video>
            <div v-else class="flex items-center justify-center h-full text-gray-600">İçerik Yok</div>
            
            <div class="absolute top-4 right-4 flex items-center gap-2 bg-black/50 px-3 py-1 rounded-full backdrop-blur-sm">
              <span :class="[screen.is_online ? 'bg-green-500 shadow-[0_0_10px_rgba(34,197,94,0.8)]' : 'bg-red-500', 'w-3 h-3 rounded-full']"></span>
              <span class="text-[10px] font-bold text-white uppercase tracking-widest">{{ screen.is_online ? 'Online' : 'Offline' }}</span>
            </div>
          </div>

          <div class="p-5 flex items-center justify-between bg-gradient-to-b from-gray-800 to-gray-900">
            <div>
              <p class="text-xs text-gray-500 uppercase font-bold tracking-tighter mb-1">Cihaz Kodu</p>
              <p class="font-mono text-xl text-purple-300">{{ screen.screen_code }}</p>
            </div>
            <div class="bg-gray-700/50 p-2 rounded-lg">
               <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
               </svg>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
const restaurant = ref(null);
const loading = ref(true);
const error = ref(null);

async function fetchRestaurantDetail() {
  const id = route.params.id;
  loading.value = true;
  error.value = null;

  try {
    // 127.0.0.1:8081'e gitmesi için nuxt.config proxy'si devrede olmalı
    const response = await fetch(`/api/restaurants/${id}`);
    
    if (!response.ok) {
      throw new Error(`Sunucu hatası: ${response.status}. Lütfen backend'i kontrol edin.`);
    }

    const data = await response.json();

    // Go'dan gelen veriyi güvenli bir yapıya oturtalım
    // Eğer Go tarafı "restaurant" ve "screens" olarak ikiye ayırmışsa direkt al
    // Ayırmamışsa, gelen veriyi bu yapıya biz sokalım
    if (data.restaurant) {
      restaurant.value = data;
    } else {
      restaurant.value = {
        restaurant: data,
        screens: data.screens || []
      };
    }

  } catch (err) {
    console.error("Detaylar çekilirken hata:", err);
    error.value = "Restoran bilgileri yüklenemedi. " + err.message;
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  fetchRestaurantDetail();
});
</script>

<style scoped>
/* Aspect ratio desteği olmayan tarayıcılar için yedek */
.aspect-video {
  aspect-ratio: 16 / 9;
}
</style>