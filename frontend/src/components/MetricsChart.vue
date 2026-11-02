<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import uPlot from "uplot";
import "uplot/dist/uPlot.min.css";
import type { MetricsSnapshot } from "../stores/metrics";

const props = defineProps<{ snapshots: MetricsSnapshot[]; field: "p99_latency_ms" | "db_qps"; color: string; label: string }>();
const chartElement = ref<HTMLElement>();
let chart: uPlot | undefined;

function render() {
  if (!chartElement.value) return;
  const data: uPlot.AlignedData = [
    props.snapshots.map((item) => new Date(item.timestamp).getTime() / 1000),
    props.snapshots.map((item) => item[props.field]),
  ];
  if (!chart) {
    chart = new uPlot({
      width: chartElement.value.clientWidth || 600,
      height: 220,
      scales: { x: { time: true } },
      series: [{}, { label: props.label, stroke: props.color, width: 2 }],
      axes: [{ stroke: "#82909d" }, { stroke: "#82909d" }],
      cursor: { drag: { x: false, y: false } },
    }, data, chartElement.value);
  } else {
    chart.setData(data);
  }
}
onMounted(render);
watch(() => props.snapshots, render, { deep: true });
onBeforeUnmount(() => chart?.destroy());
</script>
<template><div ref="chartElement" class="chart"></div></template>
<style scoped>.chart { min-width: 0; width: 100%; }</style>
