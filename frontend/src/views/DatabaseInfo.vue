<template>
  <div>
    <h1>Database Information</h1>
    <p>{{ name }}</p>
  </div>
</template>

<script setup>
import { useRoute } from 'vue-router'
const route = useRoute()
const name = route.params.name
</script> 