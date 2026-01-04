<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-900 px-4">
    <div class="bg-gray-800 p-8 rounded-xl shadow-2xl w-full max-w-md border border-gray-700">
      <div class="text-center mb-8">
        <h2 class="text-white text-3xl font-bold">Panel Girişi</h2>
        <p class="text-gray-400 mt-2">Devam etmek için şifrenizi girin</p>
      </div>
      
      <input 
        v-model="password" 
        type="password" 
        placeholder="Sistem Şifresi" 
        class="w-full p-3 rounded-lg mb-6 bg-gray-700 text-white outline-none border border-gray-600 focus:border-purple-500 transition-all"
        @keyup.enter="handleLogin"
      />
      
      <button 
        @click="handleLogin" 
        :disabled="loading"
        class="w-full bg-purple-600 text-white p-3 rounded-lg font-semibold hover:bg-purple-700 transition-all disabled:opacity-50"
      >
        {{ loading ? 'Giriş yapılıyor...' : 'Giriş Yap' }}
      </button>
    </div>
  </div>
</template>

<script setup>
definePageMeta({ layout: false }); // Login sayfasında ana menü görünmesin
const password = ref('');
const loading = ref(false);

const handleLogin = async () => {
  if (!password.value) return;
  loading.value = true;
  try {
    const res = await fetch('/api/login', {
      method: 'POST',
      body: JSON.stringify({ password: password.value })
    });
    
    if (res.ok) {
      const data = await res.json();
      localStorage.setItem('auth_token', data.token); // Şifreyi hafızaya al
      navigateTo('/');
    } else {
      alert("Şifre hatalı!");
    }
  } catch (e) {
    alert("Sunucuya bağlanılamadı!");
  } finally {
    loading.value = false;
  }
};
</script>