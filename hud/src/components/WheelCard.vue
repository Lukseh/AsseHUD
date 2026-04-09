<template>
  <div
    :style="[
      {
        width: size + 'px',
        height: size + 'px',
        backgroundImage: `url(${bg})`,
      },
      rotateStyle,
    ]"
    class="box"
  ></div>
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";

export default defineComponent({
  props: {
    value: {
      type: Number,
      default: 0,
    },
    size: {
      type: [String, Number],
      default: "100",
    },
    bg: {
      type: String,
      default: "",
    },
    maxangle: {
      type: Number,
      default: 900,
    },
  },
  setup(props) {
    const rotateStyle = computed(() => {
      const deg = props.value * (props.maxangle / 2);
      return {
        transform: `rotate(${deg}deg)`,
        transition: "transform 0.1s ease",
      };
    });
    return {
      rotateStyle,
    };
  },
});
</script>

<style scoped>
.box {
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-size: contain;
  background-repeat: no-repeat;
}
</style>
