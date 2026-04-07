<script setup lang="ts">
import HelloWorld from './components/HelloWorld.vue'
import {onMounted} from "vue";

onMounted(async () => {
  try {
    const res = await fetch('/api/hello')
    const data = await res.json()
    console.log(data)
  } catch (err) {
    console.error(err)
  }
})

</script>

<template>
  <HelloWorld />
  <p>Hello World!</p>
</template>
