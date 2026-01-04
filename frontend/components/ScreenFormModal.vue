<template>
  <div v-if="modelValue" class="modal-backdrop flex items-center justify-center" @click.self="close">
    <div class="modal-card">
      <div class="modal-header">
        <h2 class="modal-title">{{ isEditMode ? 'Ekranı Düzenle' : 'Yeni Ekran Oluştur' }}</h2>
        <button @click="close" type="button" class="modal-close-btn">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <form @submit.prevent="handleSubmit">
        <div class="grid grid-cols-1 gap-6">
          <div>
            <label class="form-label">Restoran</label>
            <select v-model="form.restaurant_id" required class="form-input">
              <option disabled value="">Lütfen bir restoran seçin</option>
              <option v-for="res in restaurants" :key="res.id" :value="res.id">{{ res.name }}</option>
            </select>
          </div>

          <div>
            <label class="form-label">Ekran Kodu</label>
            <input type="text" v-model="form.screen_code" required class="form-input" placeholder="Örn: MVL01">
          </div>

          <div>
            <label class="form-label">İçerik Tipi</label>
            <select v-model="form.content_type" required class="form-input">
              <option value="image">Resim</option>
              <option value="video">Video</option>
            </select>
          </div>

          <div>
            <label class="form-label">Medya Dosyası {{ isEditMode ? '(Değiştirmek için seçin)' : '*' }}</label>
            <input type="file" @change="handleFileChange" :required="!isEditMode" class="file-input" accept="image/*,video/*"/>
            <p v-if="isEditMode && form.media_url" class="text-xs text-gray-500 mt-1 truncate">
              Mevcut: {{ form.media_url }}
            </p>
          </div>
        </div>

        <div class="mt-6 flex justify-end space-x-4">
          <button type="button" @click="close" class="btn btn-secondary">İptal</button>
          <button type="submit" :disabled="isSubmitting" class="btn btn-primary disabled:opacity-50">
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
    const response = await fetch('/api/restaurants');
    if (!response.ok) throw new Error('Restoranlar çekilemedi');
    restaurants.value = await response.json();
  } catch (error) {
    console.error("Restoran listesi hatası:", error);
  }
}

watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    if (props.screenToEdit) {
      form.value = { ...props.screenToEdit };
    } else {
      form.value = { restaurant_id: '', screen_code: '', content_type: 'image', media_url: '' };
    }
    selectedFile.value = null;
  }
});

const handleFileChange = (event) => {
  selectedFile.value = event.target.files[0];
};

const handleSubmit = async () => {
  isSubmitting.value = true;
  
  const formData = new FormData();
  formData.append('restaurant_id', String(form.value.restaurant_id));
  formData.append('screen_code', form.value.screen_code);
  formData.append('content_type', form.value.content_type);
  
  if (selectedFile.value) {
    formData.append('media_file', selectedFile.value);
  }

  try {
    const url = isEditMode.value ? `/api/screens/${form.value.id}` : '/api/screens';
    const response = await fetch(url, {
      method: 'POST', 
      body: formData 
    });

    if (!response.ok) {
      const errorMsg = await response.text();
      throw new Error(errorMsg || 'Sunucu hatası oluştu.');
    }

    emit('saved');
    close();
  } catch (error) {
    console.error("Hata:", error);
    alert(`Hata: ${error.message}`);
  } finally {
    isSubmitting.value = false;
  }
};

const close = () => {
  emit('update:modelValue', false);
};

onMounted(fetchRestaurants);
</script>