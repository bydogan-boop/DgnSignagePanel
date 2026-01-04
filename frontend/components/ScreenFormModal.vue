<template>
  <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
    <div class="bg-gray-800 border border-gray-700 rounded-xl shadow-2xl w-full max-w-md overflow-hidden">
      
      <div class="p-6 border-b border-gray-700 flex justify-between items-center bg-gray-800/50">
        <h3 class="text-xl font-bold text-white">
          {{ isEditMode ? 'Ekranı Düzenle' : 'Yeni Ekran Ekle' }}
        </h3>
        <button @click="close" class="text-gray-400 hover:text-white transition-colors">&times;</button>
      </div>

      <form @submit.prevent="handleSubmit" class="p-6 space-y-4">
        
        <div>
          <label class="block text-sm font-medium text-gray-400 mb-1">Bağlı Restoran</label>
          <select v-model="form.restaurant_id" required class="w-full bg-gray-900 text-white p-3 rounded-lg border border-gray-600 focus:border-purple-500 outline-none">
            <option value="" disabled>Restoran Seçin</option>
            <option v-for="res in restaurants" :key="res.id" :value="res.id">{{ res.name }}</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-400 mb-1">Ekran Kodu (Benzersiz)</label>
          <input v-model="form.screen_code" type="text" placeholder="Örn: EKRAN-01" required
            class="w-full bg-gray-900 text-white p-3 rounded-lg border border-gray-600 focus:border-purple-500 outline-none uppercase" />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-400 mb-1">İçerik Tipi</label>
          <div class="flex gap-4">
            <label class="flex items-center text-white cursor-pointer">
              <input type="radio" v-model="form.content_type" value="image" class="mr-2 text-purple-600" /> Resim
            </label>
            <label class="flex items-center text-white cursor-pointer">
              <input type="radio" v-model="form.content_type" value="video" class="mr-2 text-purple-600" /> Video
            </label>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-400 mb-1">Medya Dosyası (Yeni Seçilmezse Eskisi Kalır)</label>
          <input type="file" @change="e => selectedFile = e.target.files[0]" :accept="form.content_type === 'image' ? 'image/*' : 'video/*'"
            class="w-full text-gray-400 text-sm file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-purple-600 file:text-white hover:file:bg-purple-700" />
        </div>

        <div class="flex gap-3 pt-4">
          <button type="button" @click="close" class="flex-1 px-4 py-3 rounded-lg bg-gray-700 text-white font-semibold hover:bg-gray-600 transition-all">İptal</button>
          <button type="submit" :disabled="isSubmitting" 
            class="flex-1 px-4 py-3 rounded-lg bg-purple-600 text-white font-semibold hover:bg-purple-700 transition-all disabled:opacity-50">
            {{ isSubmitting ? 'Kaydediliyor...' : 'Kaydet' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue';

const props = defineProps({
  modelValue: { type: Boolean, required: true },
  screenToEdit: { type: Object, default: null },
});

const emit = defineEmits(['update:modelValue', 'saved']);
const { fetchWithAuth } = useApi();

const restaurants = ref([]);
const isSubmitting = ref(false);
const selectedFile = ref(null);
const form = ref({
  restaurant_id: '',
  screen_code: '',
  content_type: 'image',
  media_url: ''
});

const isEditMode = computed(() => !!props.screenToEdit);

async function fetchRestaurants() {
  try {
    const response = await fetchWithAuth('/api/restaurants');
    if (response.ok) {
      restaurants.value = await response.json();
    }
  } catch (error) {
    console.error("Restoran listesi hatası:", error);
  }
}

const handleSubmit = async () => {
  if (!form.value.restaurant_id || !form.value.screen_code) {
    alert("Lütfen tüm alanları doldurun.");
    return;
  }

  isSubmitting.value = true;
  const formData = new FormData();
  formData.append('restaurant_id', String(form.value.restaurant_id));
  formData.append('screen_code', form.value.screen_code);
  formData.append('content_type', form.value.content_type);
  
  if (selectedFile.value) {
    formData.append('media_file', selectedFile.value);
  }

  try {
    // ID kontrolü: Düzenleme modunda id'yi formdan al
    const url = isEditMode.value ? `/api/screens/${props.screenToEdit.id}` : '/api/screens';
    
    const response = await fetchWithAuth(url, {
      method: 'POST',
      body: formData,
      // ÖNEMLİ: FormData gönderirken Content-Type header'ını silmeliyiz ki tarayıcı otomatik ayarlasın
      headers: { 'Content-Type': 'DELETE' } 
    });

    if (!response.ok) throw new Error('Sunucu hatası');

    emit('saved');
    close();
  } catch (error) {
    alert(`Hata: ${error.message}`);
  } finally {
    isSubmitting.value = false;
  }
};

const close = () => {
  emit('update:modelValue', false);
};

watch(() => props.modelValue, (val) => {
  if (val) {
    fetchRestaurants();
    if (props.screenToEdit) {
      form.value = { ...props.screenToEdit };
    } else {
      form.value = { restaurant_id: '', screen_code: '', content_type: 'image', media_url: '' };
    }
    selectedFile.value = null;
  }
});

onMounted(fetchRestaurants);
</script>