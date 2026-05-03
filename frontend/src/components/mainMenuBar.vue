<script setup lang="ts">
import InputText from 'primevue/inputtext';
import Button from 'primevue/button';
import Menu from 'primevue/menu';
import { ref, onMounted, onUnmounted } from 'vue';
import 'primeicons/primeicons.css';
import { useRouter } from 'vue-router'

const router = useRouter();

const isMobile = ref(window.innerWidth < 768);

const onResize = () => isMobile.value = window.innerWidth < 768;
onMounted(() => window.addEventListener('resize', onResize));
onUnmounted(() => window.removeEventListener('resize', onResize));

const searchQuery = ref('');
const menu = ref();

const userMenuItems = ref([
  {
    label: 'Profile',
    icon: 'pi pi-user',
    command: () => router.push('/profile'),
  },
  {
    label: 'Settings',
    icon: 'pi pi-cog',
    command: () => router.push('/settings'),
  },
  {
    separator: true,
  },
  {
    label: 'Logout',
    icon: 'pi pi-sign-out',
    command: () => logout(),
  },
]);

const toggleMenu = (event: Event) => {
  menu.value.toggle(event);
};

async function logout() {
  await fetch(`${import.meta.env.VITE_API_URL}/api/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  })
  await router.push('/?logout=true')
}
</script>

<template>
  <div class="flex items-center gap-2 px-4 py-2">

    <!-- Left -->
    <div class="flex items-center gap-2 flex-1">
      <Button
          :label="isMobile ? undefined : 'New Record'"
          icon="pi pi-plus"
          @click="router.push('/new-record')"
      />
      <Button
          :label="isMobile ? undefined : 'Home'"
          icon="pi pi-home"
          @click="router.push('/dashboard')"
      />
    </div>

    <!-- Centre -->
    <div class="flex-1 min-w-0">
      <InputText v-model="searchQuery" placeholder="Search..." class="w-full" />
    </div>

    <!-- Right -->
    <div class="flex items-center gap-2 flex-1 justify-end">
      <Button
          :label="isMobile ? undefined : 'Support'"
          icon="pi pi-question-circle"
          @click="router.push('/support')"
      />
      <Button
          :label="isMobile ? undefined : 'Options'"
          icon="pi pi-user"
          @click="toggleMenu"
          aria-haspopup="true"
          aria-controls="user-menu"
      />
      <Menu id="user-menu" ref="menu" :model="userMenuItems" :popup="true" />
    </div>

  </div>
</template>

<style scoped>
</style>