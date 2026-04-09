<script setup lang="ts">
import { useWebSocket } from "@vueuse/core";
import { computed, ref, onMounted } from "vue";
import WheelCard from "./components/WheelCard.vue";
import GraphCard from "./components/GraphCard.vue";
import type { PhysicsData, StaticData, AppConfig } from "./types";

const cfg = ref<AppConfig | null>(null);
const element = ref<string | undefined>(undefined);
const text = ref<boolean>(true);
const icon = ref<boolean>(true);
const all = computed(() => !element.value);

const elements = computed<string[]>(() =>
  element.value ? element.value.split(";") : [],
);

onMounted(async () => {
  const params = new URLSearchParams(window.location.search);
  element.value = params.get("element") ?? undefined;

  const textParam = params.get("text");
  if (textParam !== null) {
    text.value = textParam !== "false";
  }

  const iconParam = params.get("icon");
  if (iconParam !== null) {
    icon.value = iconParam !== "false";
  }

  try {
    const response = await fetch("/config.json");
    if (!response.ok) {
      throw new Error(
        `fetch /config.json failed: ${response.status} ${response.statusText}`,
      );
    }
    const data = (await response.json()) as AppConfig;
    cfg.value = data;

    if (textParam === null) {
      text.value = cfg.value.page.default_text;
    }
    if (iconParam === null) {
      icon.value = cfg.value.page.default_icon;
    }
  } catch (e) {
    console.error("Config load failed:", e);
  }
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
        :value="physics?.steering ?? 0"
        :maxangle="cfg?.page.wheel_max_angle"
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
    <div v-if="all || elements?.includes('brakegraph')">
      <span v-if="icon" class="mdi mdi-alert-circle-outline" /><span v-if="text"
        >Brake Graph</span
      >
      <GraphCard
        color="red"
        :max_values="
          ((cfg?.page.graph_duration ?? 5) * 1000) /
          (cfg?.polling_interval_ms ?? 30)
        "
        :value="physics?.brake"
        :poll="cfg?.polling_interval_ms"
        class="transition-none"
      />
    </div>
    <div v-if="all || elements?.includes('gasgraph')">
      <span v-if="icon" class="mdi mdi-speedometer" /><span v-if="text"
        >Gas Graph</span
      >
      <GraphCard
        color="green"
        :max_values="
          ((cfg?.page.graph_duration ?? 5) * 1000) /
          (cfg?.polling_interval_ms ?? 30)
        "
        :value="physics?.gas"
        :poll="cfg?.polling_interval_ms"
      />
    </div>
  </div>
</template>
