<template>
  <div v-if="modelValue" class="modal-backdrop flex items-center justify-center" @click.self="close">
    <div class="modal-card">
      <div class="modal-header">
        <h2 class="modal-title">{{ isEditMode ? 'Restoranı Düzenle' : 'Yeni Restoran Oluştur' }}</h2>
        <button @click="close" type="button" class="modal-close-btn" aria-label="Close modal">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <form @submit.prevent="handleSubmit">
        <div class="space-y-6">
          <div>
            <label for="restaurant_name" class="form-label">Restoran Adı</label>
            <input type="text" id="restaurant_name" v-model="form.name" required class="form-input">
          </div>
        </div>

        <div class="mt-8 flex justify-end space-x-4">
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
import { ref, watch, computed } from 'vue';

const props = defineProps({
  modelValue: { type: Boolean, required: true },
  restaurantToEdit: { type: Object, default: null },
});

const emit = defineEmits(['update:modelValue', 'saved']);

// 1. Şifreli API aracını içeri al
const { fetchWithAuth } = useApi();

const form = ref({ name: '' });
const isSubmitting = ref(false);

const isEditMode = computed(() => !!props.restaurantToEdit);

watch(() => props.restaurantToEdit, (newVal) => {
  if (newVal) {
    form.value = { name: newVal.name };
  } else {
    form.value = { name: '' };
  }
}, { immediate: true });

const handleSubmit = async () => {
  if (!form.value.name) {
    alert('Restoran adı boş olamaz.');
    return;
  }
  isSubmitting.value = true;

  try {
    const url = isEditMode.value ? `/api/restaurants/${props.restaurantToEdit.id}` : '/api/restaurants';
    
    // 2. Go tarafı POST beklediği için ikisini de POST yapıyoruz (veya Go route'a göre method seçin)
    // Şifreli fetch kullanıyoruz
    const response = await fetchWithAuth(url, {
      method: 'POST',
      body: JSON.stringify({ name: form.value.name }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || 'Bir hata oluştu.');
    }

    emit('saved');
    close();
  } catch (error) {
    console.error("Form gönderilirken hata:", error);
    alert(`Hata: ${error.message}`);
  } finally {
    isSubmitting.value = false;
  }
};

const close = () => {
  emit('update:modelValue', false);
};
</script>
