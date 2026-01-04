<template>
  <div class="container mx-auto p-4 md:p-6 space-y-8">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-white">Ekran Yönetimi</h1>
      <button @click="openCreateModal" class="btn btn-primary">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
        </svg>
        Yeni Ekran Ekle
      </button>
    </div>

    <div v-if="loading" class="text-center py-10 text-gray-400">Yükleniyor...</div>

    <div v-else class="table-wrapper overflow-x-auto">
      <table class="min-w-full">
        <thead>
          <tr>
            <th class="table-header">Önizleme</th>
            <th class="table-header">Restoran</th>
            <th class="table-header">Ekran Kodu</th>
            <th class="table-header">İçerik Tipi</th>
            <th class="table-header">Durum</th>
            <th class="table-header">Son Görülme</th>
            <th class="table-header text-right">Eylemler</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-700">
          <tr v-for="screen in screens" :key="screen.id" class="table-row">
            <td class="table-cell">
              <div class="w-20 h-12 bg-gray-900 rounded overflow-hidden border border-gray-700 flex items-center justify-center shadow-inner">
                <img v-if="screen.content_type === 'image' && screen.media_url" 
                    :src="screen.media_url" 
                    class="w-full h-full object-cover" />
                
                <video v-else-if="screen.content_type === 'video' && screen.media_url" 
                      :src="screen.media_url" 
                      class="w-full h-full object-cover"
                      autoplay 
                      muted 
                      loop 
                      playsinline>
                </video>
                
                <span v-else class="text-gray-600 text-[10px] font-bold">YOK</span>
              </div>
            </td>
            
            <td class="table-cell font-medium text-gray-200">
              {{ screen.restaurant_name }}
            </td>

            <td class="table-cell font-mono text-purple-400 font-bold">
              {{ screen.screen_code }}
            </td>

            <td class="table-cell uppercase text-[10px]">
              <span class="bg-gray-700 px-2 py-1 rounded text-gray-300 font-bold">
                {{ screen.content_type }}
              </span>
            </td>

            <td class="table-cell">
              <span :class="[screen.is_online ? 'bg-green-500' : 'bg-red-500', 'px-2 py-0.5 text-[10px] font-bold text-white rounded-full uppercase']">
                {{ screen.is_online ? 'Aktif' : 'Pasif' }}
              </span>
            </td>

            <td class="table-cell text-gray-400 text-sm">
              {{ formatLastSeen(screen.last_seen_at) }}
            </td>

            <td class="table-cell text-right space-x-4">
              <button @click="openEditModal(screen)" class="text-indigo-400 hover:text-indigo-300 font-semibold">Düzenle</button>
              <button @click="deleteScreen(screen.id)" class="text-red-500 hover:text-red-400 font-semibold">Sil</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <ScreenFormModal 
      v-model="isModalOpen" 
      :screen-to-edit="selectedScreen" 
      @saved="fetchScreens" 
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';

const { fetchWithAuth } = useApi();

const screens = ref([]);
const loading = ref(true);
const isModalOpen = ref(false);
const selectedScreen = ref(null);

async function fetchScreens() {
  loading.value = true;
  try {
    const response = await fetchWithAuth('/api/screens');
    if (!response.ok) throw new Error('Ekranlar yüklenemedi');
    screens.value = await response.json();
  } catch (error) {
    console.error("Hata:", error);
  } finally {
    loading.value = false;
  }
}

async function deleteScreen(id) {
  if (!confirm('Bu ekranı silmek istediğinize emin misiniz?')) return;
  try {
    const response = await fetchWithAuth(`/api/screens/${id}`, { method: 'DELETE' });
    if (response.ok) fetchScreens();
  } catch (error) {
    alert('Silme işlemi başarısız');
  }
}

function formatLastSeen(lastSeen) {
  if (!lastSeen || !lastSeen.Valid) return '-';
  const date = new Date(lastSeen.Time);
  // Daha kısa ve okunur bir tarih formatı
  return date.toLocaleTimeString('tr-TR', { hour: '2-digit', minute: '2-digit' }) + ' ' + date.toLocaleDateString('tr-TR', { day: '2-digit', month: '2-digit' });
}

function openCreateModal() {
  selectedScreen.value = null;
  isModalOpen.value = true;
}

function openEditModal(screen) {
  selectedScreen.value = screen;
  isModalOpen.value = true;
}

onMounted(fetchScreens);
</script>