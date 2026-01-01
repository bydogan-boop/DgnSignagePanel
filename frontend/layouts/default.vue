<template>
  <div>
    <header class="bg-gray-800 text-white p-4">
      <nav class="container mx-auto flex justify-between">
        <h1 class="text-xl font-bold">Signage Yönetim Paneli</h1>
        <div>
          <NuxtLink to="/" class="px-4">Ana Sayfa</NuxtLink>
          <NuxtLink to="/restaurants" class="px-4">Restoranlar</NuxtLink>
          <NuxtLink to="/screens" class="px-4">Ekranlar</NuxtLink>
        </div>
      </nav>
    </header>
    <main class="container mx-auto p-4">
      <slot />
    </main>
  </div>
</template>

<style>
/* Global stiller ve Tailwind CSS importları buraya gelebilir */
@import '~/assets/css/main.css';
</style>
