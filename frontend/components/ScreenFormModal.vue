<script setup>
import { ref, watch, onMounted, computed } from 'vue';

const props = defineProps({
  modelValue: { type: Boolean, required: true },
  screenToEdit: { type: Object, default: null },
});

const emit = defineEmits(['update:modelValue', 'saved']);

// Şifreli API fonksiyonu
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
    // fetch -> fetchWithAuth
    const response = await fetchWithAuth('/api/restaurants');
    if (response.ok) {
      restaurants.value = await response.json();
    }
  } catch (error) {
    console.error("Restoran listesi hatası:", error);
  }
}

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
    
    // fetch -> fetchWithAuth (Dosya gönderimi otomatik olarak Auth header ile gidecek)
    const response = await fetchWithAuth(url, {
      method: 'POST', 
      body: formData 
    });

    if (!response.ok) {
      const errorMsg = await response.text();
      throw new Error(errorMsg || 'Sunucu hatası');
    }

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
    if (props.screenToEdit) form.value = { ...props.screenToEdit };
    else form.value = { restaurant_id: '', screen_code: '', content_type: 'image', media_url: '' };
    selectedFile.value = null;
  }
});

onMounted(fetchRestaurants);
</script>