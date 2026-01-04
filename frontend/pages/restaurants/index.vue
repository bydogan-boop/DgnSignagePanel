<template>
  <div class="container mx-auto p-4 md:p-6">
    <!-- Sayfa Başlığı ve Eylem Butonu -->
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-title">Tüm Restoranlar</h1>
      <button @click="openModal()" class="btn btn-primary">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
        </svg>
        Yeni Restoran Ekle
      </button>
    </div>

    <!-- Yüklenme ve Boş Durum Göstergeleri -->
    <div v-if="loading" class="text-center py-10 text-gray-400">Yükleniyor...</div>
    <div v-else-if="!loading && restaurants.length === 0" class="text-center py-10 text-gray-400 bg-gray-800 rounded-2xl">
      Gösterilecek restoran bulunmuyor.
    </div>

    <!-- Restoran Tablosu -->
    <div v-else class="table-wrapper overflow-x-auto">
        <table class="min-w-full">
            <thead>
                <tr>
                    <th scope="col" class="table-header">Logo</th>
                    <th scope="col" class="table-header">Restoran Adı</th>
                    <th scope="col" class="table-header">Adres</th>
                    <th scope="col" class="table-header">Ekran Sayısı</th>
                    <th scope="col" class="table-header">
                        <span class="sr-only">Eylemler</span>
                    </th>
                </tr>
            </thead>
            <tbody class="divide-y divide-gray-700">
                <tr v-for="restaurant in restaurants" :key="restaurant.id" class="table-row">
                    <td class="table-cell">
                        <img :src="restaurant.logo_url || '/placeholder-logo.svg'" alt="Logo" class="w-12 h-12 rounded-full object-cover border-2 border-gray-600">
                    </td>
                    <td class="table-cell">
                        <router-link :to="`/restaurants/${restaurant.id}`" class="text-lg font-medium text-purple-400 hover:text-purple-300">
                            {{ restaurant.name }}
                        </router-link>
                        <span :class="[restaurant.is_any_screen_online ? 'bg-green-500' : 'bg-red-500', 'ml-3 px-2 py-0.5 text-xs font-semibold text-white rounded-full align-middle']">
                            {{ restaurant.is_any_screen_online ? 'Aktif' : 'Pasif' }}
                        </span>
                    </td>
                    <td class="table-cell max-w-xs truncate">
                        {{ restaurant.address }}
                    </td>
                    <td class="table-cell text-center">
                        {{ restaurant.screen_count }}
                    </td>
                    <td class="table-cell text-right space-x-4">
                        <button @click="openModal(restaurant)" class="text-indigo-400 hover:text-indigo-300 font-semibold">Düzenle</button>
                        <button @click="confirmDelete(restaurant.id)" class="text-red-500 hover:text-red-400 font-semibold">Sil</button>
                    </td>
                </tr>
            </tbody>
        </table>
    </div>

    <!-- Modal (Değişiklik Yok) -->
    <div v-if="showModal" class="modal-backdrop flex items-center justify-center" @click.self="closeModal">
      <div class="modal-card m-4">
        <div class="p-6">
          <div class="modal-header">
              <h3 class="modal-title">{{ form.id ? 'Restoranı Düzenle' : 'Yeni Restoran Oluştur' }}</h3>
              <button @click="closeModal" type="button" class="modal-close-btn">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
          </div>
          
          <form @submit.prevent="saveRestaurant">
            <div class="grid grid-cols-1 gap-6">
                <div class="flex items-center space-x-6">
                    <div class="shrink-0">
                        <img class="h-24 w-24 object-cover rounded-full border-2 border-gray-600" :src="logoPreview || '/placeholder-logo.svg'" alt="Mevcut logo" />
                    </div>
                    <label class="block">
                        <span class="sr-only">Logo seç</span>
                        <input type="file" @change="handleLogoChange" accept="image/*" class="file-input"/>
                    </label>
                </div>
              <input type="text" v-model="form.name" placeholder="Restoran Adı *" required class="form-input">
              <input type="tel" v-model="form.phone" placeholder="Telefon Numarası" class="form-input">
              <textarea v-model="form.address" placeholder="Adres" rows="3" class="form-input"></textarea>
            </div>
            <div class="flex justify-end space-x-4 mt-8">
              <button type="button" @click="closeModal" class="btn btn-secondary">İptal</button>
              <button type="submit" :disabled="isSaving" class="btn btn-primary disabled:opacity-50">
                {{ isSaving ? 'Kaydediliyor...' : 'Kaydet' }}
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
import { useRouter } from 'vue-router';

const restaurants = ref([]);
const loading = ref(true);
const showModal = ref(false);
const isSaving = ref(false);
const router = useRouter();

const initialFormState = {
  id: null,
  name: '',
  phone: '',
  address: '',
  logo: null,
};
const form = ref({ ...initialFormState });
const logoPreview = ref(null);

async function fetchRestaurants() {
  loading.value = true;
  try {
    const response = await fetch('/api/restaurants/status');
    if (!response.ok) throw new Error('Restoranlar alınamadı');
    restaurants.value = await response.json();
  } catch (error) { 
    console.error("Restoranlar çekilirken hata:", error);
    restaurants.value = [];
  } finally {
    loading.value = false;
  }
}

function handleLogoChange(event) {
  const file = event.target.files[0];
  if (file) {
    form.value.logo = file;
    const reader = new FileReader();
    reader.onload = (e) => {
      logoPreview.value = e.target.result;
    };
    reader.readAsDataURL(file);
  }
}

function openModal(restaurant = null) {
  if (restaurant) {
    form.value.id = restaurant.id;
    form.value.name = restaurant.name;
    form.value.phone = restaurant.phone;
    form.value.address = restaurant.address;
    logoPreview.value = restaurant.logo_url;
  } else {
    Object.assign(form.value, initialFormState);
    logoPreview.value = null;
  }
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
  Object.assign(form.value, initialFormState);
  logoPreview.value = null;
  const fileInput = document.querySelector('input[type="file"]');
  if(fileInput) fileInput.value = '';
}

async function saveRestaurant() {
  isSaving.value = true;

  const formData = new FormData();
  formData.append('name', form.value.name);
  formData.append('phone', form.value.phone);
  formData.append('address', form.value.address);
  if (form.value.logo instanceof File) {
    formData.append('logo', form.value.logo);
  }

  const isUpdating = !!form.value.id;
  const url = isUpdating ? `/api/restaurants/${form.value.id}` : '/api/restaurants';
  const method = isUpdating ? 'PUT' : 'POST';

  try {
    const response = await fetch(url, { method, body: formData });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || `Restoran ${isUpdating ? 'güncellenemedi' : 'oluşturulamadı'}.`);
    }
    
    await fetchRestaurants();
    closeModal();

  } catch (error) {
    console.error('Kaydetme hatası:', error);
    alert(`Hata: ${error.message}`);
  } finally {
    isSaving.value = false;
  }
}

async function confirmDelete(id) {
  if (!confirm('Bu restoranı ve ilişkili tüm verileri (ekranlar, logolar vb.) kalıcı olarak silmek istediğinizden emin misiniz? Bu işlem geri alınamaz.')) {
    return;
  }
  try {
    const response = await fetch(`/api/restaurants/${id}`, { method: 'DELETE' });
    if (!response.ok) {
        const err = await response.json();
        throw new Error(err.message || 'Restoran silinemedi.');
    }
    await fetchRestaurants();
  } catch (error) {
    console.error('Restoran silinirken bir hata oluştu:', error);
    alert(`Hata: ${error.message}`);
  }
}

onMounted(fetchRestaurants);

</script>