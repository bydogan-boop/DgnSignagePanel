<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-3xl font-bold">Ekranlar</h2>
      <button @click="openModal" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
        Yeni Ekran Ekle
      </button>
    </div>

    <!-- Ekranlar Tablosu -->
    <div class="bg-white shadow-md rounded-lg overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Görüntü</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Restoran Adı</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Ekran Kodu</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">İçerik Türü</th>
            <th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Eylemler</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <!-- TODO: API'den gelen ekran verileri buraya gelecek -->
          <tr v-if="screens.length === 0">
            <td colspan="5" class="px-6 py-4 text-center text-gray-500">Henüz ekran eklenmemiş.</td>
          </tr>
          <!-- Örnek bir satır -->
          <tr v-for="screen in screens" :key="screen.id">
            <td class="px-6 py-4 whitespace-nowrap">
              <img v-if="screen.content_type === 'image'" :src="screen.media_url" alt="Ekran Görüntüsü" class="h-10 w-16 object-cover rounded">
               <video v-else-if="screen.content_type === 'video'" :src="screen.media_url" class="h-10 w-16 object-cover rounded" muted loop></video>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ screen.restaurant_name }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ screen.screen_code }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ screen.content_type }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <button class="text-indigo-600 hover:text-indigo-900">Düzenle</button>
              <button class="text-red-600 hover:text-red-900 ml-4">Sil</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Yeni Ekran Ekleme Modalı -->
    <div v-if="showModal" class="fixed z-10 inset-0 overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div class="fixed inset-0 transition-opacity" aria-hidden="true">
          <div class="absolute inset-0 bg-gray-500 opacity-75"></div>
        </div>
        <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
        <div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <form @submit.prevent="addScreen">
            <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
              <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
                Yeni Ekran Oluştur
              </h3>
              <div class="mt-4 space-y-4">
                <div>
                  <label for="restaurant" class="block text-sm font-medium text-gray-700">Restoran</label>
                  <!-- TODO: Restoranlar API'den yüklenecek -->
                  <select id="restaurant" v-model="newScreen.restaurant_id" class="mt-1 block w-full py-2 px-3 border border-gray-300 bg-white rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                    <option value="">Restoran Seçin</option>
                  </select>
                </div>
                <div>
                  <label for="screen_code" class="block text-sm font-medium text-gray-700">Benzersiz Ekran Kodu</label>
                  <input type="text" v-model="newScreen.screen_code" id="screen_code" class="mt-1 focus:ring-indigo-500 focus:border-indigo-500 block w-full shadow-sm sm:text-sm border-gray-300 rounded-md">
                </div>
                 <div>
                  <label for="content_type" class="block text-sm font-medium text-gray-700">İçerik Türü</label>
                  <select id="content_type" v-model="newScreen.content_type" class="mt-1 block w-full py-2 px-3 border border-gray-300 bg-white rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                    <option value="video">Video</option>
                    <option value="image">Resim</option>
                  </select>
                </div>
                <div>
                  <label for="media_file" class="block text-sm font-medium text-gray-700">İçerik Dosyası</label>
                  <input type="file" @change="onFileChange" id="media_file" class="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100">
                </div>
              </div>
            </div>
            <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
              <button type="submit" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none sm:ml-3 sm:w-auto sm:text-sm">
                Ekranı Oluştur
              </button>
              <button @click="closeModal" type="button" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">
                İptal
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';

const screens = ref([]); // API'den gelecek ekranların listesi
const showModal = ref(false);
const newScreen = ref({
  restaurant_id: '',
  screen_code: '',
  content_type: 'video',
  media_file: null,
});

// TODO: API'den ekranları ve restoranları çekmek için onMounted kullanılacak
onMounted(async () => {
  // fetchScreens();
  // fetchRestaurants();
});

function onFileChange(e) {
  newScreen.value.media_file = e.target.files[0];
}

function openModal() {
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
}

async function addScreen() {
  // TODO: FormData kullanarak API'ye POST isteği gönderilecek
  console.log('Yeni ekran bilgileri:', newScreen.value);
  
  // İstek sonrası modalı kapat
  closeModal();
}
</script>
