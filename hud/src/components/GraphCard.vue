<template>
  <v-sparkline
    color="white"
    height="80"
    :min="0"
    :max="1"
    :model-value="values"
    :smooth="false"
  />
</template>
<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
const props = defineProps({
  max_values: {
    type: Number,
    default: 50,
  },
  value: {
    type: Number,
    default: 0,
  },
  poll: {
    type: Number,
    default: 30,
  },
});

function getMaxValues() {
  const n = Number(props.max_values);
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : 50;
}

const values = ref<number[]>([]);

function addValue(newValue: number) {
  const maxValues = getMaxValues();
  const nextValues = [...values.value, newValue];
  if (nextValues.length > maxValues) {
    nextValues.shift();
  }
  values.value = nextValues;
}

let interval: number | undefined;

onMounted(() => {
  values.value = Array(getMaxValues()).fill(0);
  interval = window.setInterval(() => {
    addValue(props.value);
  }, props.poll);
});

onUnmounted(() => {
  if (interval !== undefined) {
    window.clearInterval(interval);
  }
});
</script>
