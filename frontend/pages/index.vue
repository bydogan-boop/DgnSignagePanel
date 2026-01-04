<template>
  <div class="container mx-auto p-4 md:p-6 space-y-8">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-white">Genel Bakış</h1>
    </div>

    <div v-if="loading" class="text-center py-10 text-gray-400">Yükleniyor...</div>
    <div v-else-if="error" class="text-center py-10 text-red-500 bg-gray-800 rounded-2xl border border-red-900/20">
      <p>{{ error }}</p>
      <button @click="fetchData" class="mt-4 text-purple-400 font-semibold hover:underline">Tekrar Dene</button>
    </div>

    <div v-else class="space-y-8">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="bg-gray-800 p-5 rounded-xl border border-gray-700 shadow-sm">
          <div class="text-gray-400 text-xs uppercase font-bold tracking-wider mb-1">Toplam Restoran</div>
          <div class="text-2xl font-bold text-white">{{ stats.restaurant_count || 0 }}</div>
        </div>

        <div class="bg-gray-800 p-5 rounded-xl border border-gray-700 shadow-sm">
          <div class="text-gray-400 text-xs uppercase font-bold tracking-wider mb-1">Toplam Ekran</div>
          <div class="text-2xl font-bold text-white">{{ stats.screen_count || 0 }}</div>
        </div>

        <div class="bg-gray-800 p-5 rounded-xl border border-gray-700 shadow-sm">
          <div class="text-gray-400 text-xs uppercase font-bold tracking-wider mb-1">Aktif Ekran</div>
          <div class="text-2xl font-bold text-green-400">{{ stats.active_screens || 0 }}</div>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        <div class="lg:col-span-2">
          <h2 class="text-lg font-bold text-gray-200 mb-4">Restoran Durumları</h2>
          <div class="bg-gray-800 rounded-xl border border-gray-700 overflow-hidden shadow-md">
            <table class="min-w-full">
              <thead>
                <tr class="bg-gray-900/50">
                  <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">Restoran Adı</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider text-center">Ekran Sayısı</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">Durum</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-700">
                <tr v-for="item in restaurantStatus" :key="item.id" class="hover:bg-gray-700/30 transition-colors">
                  <td class="px-4 py-4 whitespace-nowrap text-sm font-medium text-purple-400">
                    {{ item.name }}
                  </td>
                  <td class="px-4 py-4 whitespace-nowrap text-sm text-gray-300 text-center">
                    {{ item.screen_count || 0 }}
                  </td>
                  <td class="px-4 py-4 whitespace-nowrap">
                    <span :class="[item.is_any_screen_online ? 'bg-green-500' : 'bg-red-500', 'px-2 py-0.5 text-[10px] font-bold text-white rounded-full uppercase']">
                      {{ item.is_any_screen_online ? 'Aktif' : 'Pasif' }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div>
          <h2 class="text-lg font-bold text-gray-200 mb-4">Son Aktif Ekranlar</h2>
          <div class="bg-gray-800 rounded-xl border border-gray-800 p-5 space-y-4 shadow-md">
            <div v-for="screen in recentScreens" :key="screen.screen_code" class="flex justify-between items-center group border-b border-gray-700/50 pb-3 last:border-0 last:pb-0">
              <div>
                <div class="text-sm font-bold text-white font-mono">{{ screen.screen_code }}</div>
                <div class="text-[11px] text-gray-500">{{ screen.restaurant_name }}</div>
              </div>
              <div class="text-[10px] text-gray-400 bg-gray-900 px-2 py-0.5 rounded border border-gray-700">
                {{ formatLastSeen(screen.last_seen_at) }}
              </div>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';

// Şifreli fetch fonksiyonumuzu çağırıyoruz
const { fetchWithAuth } = useApi();

const stats = ref({});
const restaurantStatus = ref([]);
const recentScreens = ref([]);
const loading = ref(true);
const error = ref(null);

async function fetchData() {
  loading.value = true;
  error.value = null;
  try {
    // Standart fetch yerine fetchWithAuth kullanıyoruz
    const response = await fetchWithAuth('/api/dashboard');
    if (!response.ok) throw new Error('Veriler alınamadı. Lütfen giriş yapın.');
    const data = await response.json();

    stats.value = {
      restaurant_count: data.restaurant_count,
      screen_count: data.screen_count,
      active_screens: data.active_screens,
    };
    restaurantStatus.value = data.restaurant_status || [];
    recentScreens.value = data.recent_screens || [];

  } catch (err) {
    console.error("Dashboard hatası:", err);
    error.value = err.message || "Veriler yüklenirken bir hata oluştu.";
  } finally {
    loading.value = false;
  }
}

function formatLastSeen(lastSeen) {
  if (!lastSeen || !lastSeen.Valid) return '-';
  const date = new Date(lastSeen.Time);
  const now = new Date();
  const diff = Math.floor((now - date) / 1000);

  if (diff < 60) return `${diff} sn önce`;
  if (diff < 3600) return `${Math.floor(diff / 60)} dk önce`;
  if (diff < 8400) return `${Math.floor(diff / 3600)} sa önce`;
  return date.toLocaleDateString('tr-TR');
}

onMounted(fetchData);
</script>