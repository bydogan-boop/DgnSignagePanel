<template>
  <div class="bg-gray-900 min-h-screen text-gray-200">
    <!-- Üst Navigasyon Menüsü -->
    <header class="bg-gray-800 shadow-lg sticky top-0 z-50">
      <nav class="container mx-auto px-4 sm:px-6 lg:px-8 flex justify-between items-center h-16">
        
        <!-- Logo/Başlık -->
        <div class="flex-shrink-0">
           <NuxtLink to="/" class="flex items-center space-x-3 text-2xl font-extrabold tracking-tight text-white">
             <img src="/logo.svg" alt="Dgn Logo" class="h-8 w-8">
             <span>Dgn<span class="text-purple-400">SignagePanel</span></span>
           </NuxtLink>
        </div>
        
        <!-- Masaüstü Navigasyon Linkleri -->
        <div class="hidden md:flex items-center space-x-2">
          <NuxtLink 
            to="/" 
            class="px-4 py-2 rounded-md text-sm font-medium text-gray-300 hover:bg-gray-700 hover:text-purple-300 transition-colors duration-200"
            active-class="text-purple-400 bg-gray-700/50"
            >Ana Sayfa</NuxtLink>
          <NuxtLink 
            to="/restaurants" 
            class="px-4 py-2 rounded-md text-sm font-medium text-gray-300 hover:bg-gray-700 hover:text-purple-300 transition-colors duration-200"
            active-class="text-purple-400 bg-gray-700/50"
            >Restoranlar</NuxtLink>
          <NuxtLink 
            to="/screens" 
            class="px-4 py-2 rounded-md text-sm font-medium text-gray-300 hover:bg-gray-700 hover:text-purple-300 transition-colors duration-200"
            active-class="text-purple-400 bg-gray-700/50"
            >Ekranlar</NuxtLink>
        </div>

        <!-- Mobil Menü Butonu (Gelecekte eklenebilir) -->
        <div class="md:hidden">
          <!-- <button>Menu</button> -->
        </div>

      </nav>
    </header>

    <!-- Ana İçerik Alanı -->
    <div>
      <slot />
    </div>

  </div>
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&display=swap');
@import '~/assets/css/main.css';

body {
  background-color: #111827; /* bg-gray-900 */
  font-family: 'Inter', sans-serif;
}

/* Aktif olan linkin fontunu daha belirgin yap */
.router-link-exact-active {
    font-weight: 600;
}
</style>

