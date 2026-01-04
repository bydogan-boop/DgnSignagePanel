<script setup>
import { ref } from 'vue'

const isMenuOpen = ref(false)

const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value
}
</script>

<template>
  <div class="bg-gray-900 min-h-screen text-gray-200">
    <header class="bg-gray-800 shadow-lg sticky top-0 z-50">
      <nav class="container mx-auto px-4 sm:px-6 lg:px-8 flex justify-between items-center h-16">
        
        <div class="flex-shrink-0">
           <NuxtLink to="/" class="flex items-center space-x-3 text-2xl font-extrabold tracking-tight text-white">
             <img src="/logo.svg" alt="Dgn Logo" class="h-8 w-8">
             <span>Dgn<span class="text-purple-400">SignagePanel</span></span>
           </NuxtLink>
        </div>
        
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

        <div class="md:hidden">
          <button @click="toggleMenu" class="text-gray-300 hover:text-white focus:outline-none p-2">
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path v-if="!isMenuOpen" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

      </nav>

      <div v-show="isMenuOpen" class="md:hidden bg-gray-800 border-t border-gray-700">
        <div class="px-2 pt-2 pb-3 space-y-1 sm:px-3">
          <NuxtLink 
            to="/" 
            @click="isMenuOpen = false"
            class="block px-3 py-2 rounded-md text-base font-medium text-gray-300 hover:bg-gray-700 hover:text-purple-300"
            active-class="text-purple-400 bg-gray-700/50"
            >Ana Sayfa</NuxtLink>
          <NuxtLink 
            to="/restaurants" 
            @click="isMenuOpen = false"
            class="block px-3 py-2 rounded-md text-base font-medium text-gray-300 hover:bg-gray-700 hover:text-purple-300"
            active-class="text-purple-400 bg-gray-700/50"
            >Restoranlar</NuxtLink>
          <NuxtLink 
            to="/screens" 
            @click="isMenuOpen = false"
            class="block px-3 py-2 rounded-md text-base font-medium text-gray-300 hover:bg-gray-700 hover:text-purple-300"
            active-class="text-purple-400 bg-gray-700/50"
            >Ekranlar</NuxtLink>
        </div>
      </div>
    </header>

    <main class="container mx-auto px-4 py-6">
      <slot />
    </main>

  </div>
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&display=swap');
/* @import '~/assets/css/main.css'; */

body {
  background-color: #111827; /* bg-gray-900 */
  font-family: 'Inter', sans-serif;
}

.router-link-exact-active {
    font-weight: 600;
}
</style>