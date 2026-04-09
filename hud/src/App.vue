<script setup lang="ts">
import { useWebSocket } from "@vueuse/core";
import { computed, ref, onMounted } from "vue";
import WheelCard from "./components/WheelCard.vue";
import type { PhysicsData, StaticData, AppConfig } from "./types";

const cfg = ref<AppConfig>();
onMounted(async () => {
  cfg.value = await fetch("/config.json").then((r) => r.json());
});

onMounted(() => {
  const params = new URLSearchParams(window.location.search);
  element.value = params.get("element") ?? undefined;
  text.value = params.get("text") === "false" ? false : true;
  icon.value = params.get("icon") === "false" ? false : true;
});
const element = ref<string | undefined>(undefined);
const text = ref<boolean | undefined>(cfg?.value?.default_icon);
const icon = ref<boolean | undefined>(cfg?.value?.default_icon);
const all = computed(() => !element.value);

const elements = ref<string[]>();

onMounted(() => {
  elements.value = element?.value?.split(";");
});

// STATIC
const { data: sdata } = useWebSocket("ws://localhost:16440/ws/static", {
  onConnected(_ws) {
    console.log("Connected!");
  },
  onDisconnected(_ws, event) {
    console.log("Disconnected!", event.code);
  },
  onError(_ws, event) {
    console.error("Error:", event);
  },
});
const st = computed<StaticData | null>(() => {
  try {
    return sdata.value ? JSON.parse(sdata.value) : null;
  } catch (e) {
    console.error("JSON parse error", e);
    return null;
  }
});

// PHYSICS
const { data: pdata } = useWebSocket("ws://localhost:16440/ws/physics", {
  onConnected(_ws) {
    console.log("Connected!");
  },
  onDisconnected(_ws, event) {
    console.log("Disconnected!", event.code);
  },
  onError(_ws, event) {
    console.error("Error:", event);
  },
});
const physics = computed<PhysicsData | null>(() => {
  try {
    return pdata.value ? JSON.parse(pdata.value) : null;
  } catch (e) {
    console.error("JSON parse error", e);
    return null;
  }
});
</script>

<template>
  <div class="text-[white] flex flex-col gap-5">
    <div v-if="all || elements?.includes('gas')">
      <div v-if="text">
        <span v-if="icon" class="mdi mdi-speedometer" /><span>GAS</span>
      </div>
      <v-progress-linear
        :model-value="(physics?.gas ?? 0) * 100"
        :height="30"
        color="green"
      />
    </div>
    <div v-if="all || elements?.includes('brake')">
      <div v-if="text">
        <span v-if="icon" class="mdi mdi-alert-circle-outline" /><span
          >Brake</span
        >
      </div>
      <v-progress-linear
        :model-value="(physics?.brake ?? 0) * 100"
        :height="30"
        color="red"
      />
    </div>
    <div v-if="all || elements?.includes('steer')">
      <div v-if="text">
        <span v-if="icon" class="mdi mdi-steering" /><span>Steering</span>
      </div>
      <wheel-card
        :value="physics?.steering"
        :maxangle="cfg?.wheel_max_angle"
        size="150"
        bg="https://q4v8e3e5.rocketcdn.me/app/uploads/sites/5/2024/03/ST-FPE-1-1000x1000.png"
      />
    </div>
    <div v-if="all || elements?.includes('fuel')">
      <div v-if="text">
        <span v-if="icon" class="mdi mdi-gas-station" /><span>FUEL</span>
      </div>
      <v-progress-linear
        color="orange"
        rounded
        :height="18"
        :model-value="((physics?.fuel ?? 0) / (st?.max_fuel ?? 1)) * 100"
      />
    </div>
    <div v-if="all || elements?.includes('rpm')">
      <div v-if="text">
        <span v-if="icon" class="mdi mdi-gauge" /><span>RPM</span>
      </div>
      <v-progress-linear
        color="red"
        :height="18"
        :model-value="physics?.rpm"
        :max="st?.max_rpm"
        class="transition-none"
        style="transition: none !important"
      />
    </div>
    <div v-if="all || elements?.includes('gear')">
      <span v-if="icon" class="mdi mdi-gauge" /><span v-if="text">GEAR</span>
      <div style="font-family: digital-7" id="gear">
        {{
          (physics?.gear ?? 0) === 1
            ? "N"
            : (physics?.gear ?? 0) < 1
              ? "R"
              : (physics?.gear ?? 0) - 1
        }}
      </div>
    </div>
  </div>
</template>
