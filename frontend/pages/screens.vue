<template>
  <div class="container mx-auto p-4 md:p-6">
    <!-- Sayfa Başlığı ve Eylem Butonu -->
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-title">Tüm Ekranlar</h1>
      <button @click="openCreateModal" class="btn btn-primary">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
        </svg>
        Yeni Ekran Ekle
      </button>
    </div>

    <!-- Ekranlar Tablosu -->
    <div class="table-wrapper">
      <div class="overflow-x-auto">
        <table class="min-w-full">
          <thead>
            <tr>
              <th scope="col" class="table-header">Önizleme</th>
              <th scope="col" class="table-header">Restoran Adı</th>
              <th scope="col" class="table-header">Ekran Kodu</th>
              <th scope="col" class="table-header">Online Durumu</th>
              <th scope="col" class="table-header">Son Görülme</th>
              <th scope="col" class="table-header">
                <span class="sr-only">Eylemler</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="6" class="px-6 py-12 text-center text-gray-500">Ekranlar yükleniyor...</td>
            </tr>
            <tr v-else-if="screens.length === 0">
              <td colspan="6" class="px-6 py-12 text-center text-gray-500">Gösterilecek ekran bulunmuyor.</td>
            </tr>
            <tr v-for="screen in screens" :key="screen.id" class="table-row">
              <td class="table-cell">
                <a v-if="screen.media_url" :href="screen.media_url" target="_blank">
                  <img v-if="screen.content_type && screen.content_type.startsWith('image')" :src="screen.media_url" alt="Önizleme" class="h-10 w-16 object-cover rounded">
                  <video v-else-if="screen.content_type && screen.content_type.startsWith('video')" :src="screen.media_url" class="h-10 w-16 object-cover rounded" muted loop autoplay playsinline></video>
                  <div v-else class="h-10 w-16 rounded bg-gray-700 flex items-center justify-center text-xs text-gray-400">URL</div>
                </a>
                <div v-else class="h-10 w-16 rounded bg-gray-700 flex items-center justify-center text-xs text-gray-400">Yok</div>
              </td>
              <td class="table-cell font-medium text-white">{{ screen.restaurant_name }}</td>
              <td class="table-cell text-gray-400 font-mono">{{ screen.screen_code }}</td>
              <td class="table-cell">
                <span :class="screen.is_online ? 'bg-green-900/50 text-green-300' : 'bg-red-900/50 text-red-400'" class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full">
                  {{ screen.is_online ? 'Aktif' : 'Pasif' }}
                </span>
              </td>
              <td class="table-cell text-gray-400">{{ formatLastSeen(screen.last_seen_at) }}</td>
              <td class="table-cell text-right font-medium space-x-4">
                 <button @click="openEditModal(screen)" class="text-purple-400 hover:text-purple-300">Düzenle</button>
                 <button @click="deleteScreen(screen.id)" class="text-red-500 hover:text-red-400">Sil</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <ScreenFormModal v-model="isModalOpen" :screen-to-edit="selectedScreen" @saved="handleSave" />

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import ScreenFormModal from '@/components/ScreenFormModal.vue';

const screens = ref([]);
const loading = ref(true);
const isModalOpen = ref(false);
const selectedScreen = ref(null);


async function fetchScreens() {
  loading.value = true;
  try {
    const response = await fetch('/api/screens');
    if (!response.ok) {
      throw new Error('Ekranlar alınamadı');
    }
    screens.value = await response.json();
  } catch (error) {
    console.error("Ekranlar çekilirken hata:", error);
    // Hata durumunda kullanıcıya bilgi vermek için bir state kullanılabilir.
  } finally {
    loading.value = false;
  }
}
async function deleteScreen(id) {
  if (!confirm('Bu ekranı silmek istediğinizden emin misiniz? Bu işlem geri alınamaz.')) {
    return;
  }

  try {
    const response = await fetch(`/api/screens/${id}`, {
      method: 'DELETE',
    });

    if (!response.ok) {
      throw new Error('Ekran silinemedi.');
    }

    // Silme işlemi başarılı olursa, ekran listesini yeniden yükle
    await fetchScreens();

  } catch (error) {
    console.error('Ekran silinirken bir hata oluştu:', error);
    alert('Ekran silinirken bir hata oluştu.');
  }
}

function formatLastSeen(lastSeen) {
  if (!lastSeen || !lastSeen.Valid) {
    return 'Bilinmiyor';
  }
  const date = new Date(lastSeen.Time);
  const now = new Date();
  const diffInSeconds = Math.floor((now - date) / 1000);

  if (diffInSeconds < 60) return `${diffInSeconds} sn önce`;
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} dk önce`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} saat önce`;

  return date.toLocaleString('tr-TR');
}

function setDefaultImage(event) {
  event.target.src = '/default-logo.svg';
}

// Modal kontrol fonksiyonları
const openCreateModal = () => {
  selectedScreen.value = null;
  isModalOpen.value = true;
};

const openEditModal = (screen) => {
  selectedScreen.value = screen;
  isModalOpen.value = true;
};

const handleSave = () => {
  isModalOpen.value = false; // Modalı kapat
  fetchScreens(); // Listeyi yenile
};


onMounted(fetchScreens);
</script>
