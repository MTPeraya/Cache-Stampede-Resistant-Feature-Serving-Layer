<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";
import { useMetricsStore } from "./stores/metrics";
import MetricsChart from "./components/MetricsChart.vue";

const metrics = useMetricsStore();
const mode = ref<"naive" | "hardened">("naive");
const burstRPS = ref(200);
const burstSeconds = ref(10);
const burstKeys = ref(10);
const notice = ref("");
const entityID = ref(1);
const result = ref("No request yet");

async function setMode(next: "naive" | "hardened") {
  const response = await fetch(`/mode?value=${next}`, { method: "POST" });
  if (!response.ok) { notice.value = "Could not change mode"; return; }
  mode.value = next;
  notice.value = `${next} mode active`;
}
async function runBurst() {
  const response = await fetch(`/loadgen/burst?rps=${burstRPS.value}&seconds=${burstSeconds.value}&keys=${burstKeys.value}`, { method: "POST" });
  notice.value = response.ok ? "Burst started" : "Burst parameters rejected";
}
async function fetchFeature() {
  const response = await fetch(`/features/${entityID.value}`);
  result.value = response.ok ? JSON.stringify(await response.json(), null, 2) : `HTTP ${response.status}`;
}
onMounted(metrics.connect);
onBeforeUnmount(metrics.disconnect);
</script>

<template>
  <main>
    <header>
      <div><p class="eyebrow">STAMPede shield / phase 3</p><h1>Feature serving under pressure.</h1><p class="intro">Watch cache behavior change in real time.</p></div>
      <span class="connection" :class="{ online: metrics.connected }">{{ metrics.connected ? "LIVE" : "OFFLINE" }}</span>
    </header>
    <section class="toolbar card">
      <div><span class="label">Read path</span><div class="segmented"><button :class="{ selected: mode === 'naive' }" @click="setMode('naive')">Naive</button><button :class="{ selected: mode === 'hardened' }" @click="setMode('hardened')">Hardened</button></div></div>
      <div><span class="label">Built-in burst</span><div class="inputs"><input v-model.number="burstRPS" type="number" min="1" /><input v-model.number="burstSeconds" type="number" min="1" /><input v-model.number="burstKeys" type="number" min="1" /><button @click="runBurst">Run burst</button></div><small>RPS / seconds / keys</small></div>
      <span class="notice">{{ notice }}</span>
    </section>
    <section class="stats">
      <div class="stat card"><span>Hit rate</span><strong>{{ ((metrics.latest?.hit_rate ?? 0) * 100).toFixed(1) }}%</strong></div>
      <div class="stat card"><span>P99 latency</span><strong>{{ (metrics.latest?.p99_latency_ms ?? 0).toFixed(1) }} ms</strong></div>
      <div class="stat card"><span>DB QPS</span><strong>{{ (metrics.latest?.db_qps ?? 0).toFixed(1) }}</strong></div>
      <div class="stat card"><span>Coalesced / stale</span><strong>{{ metrics.latest?.coalesced ?? 0 }} / {{ metrics.latest?.stale_serves ?? 0 }}</strong></div>
    </section>
    <section class="charts">
      <article class="card"><h2>P99 request latency</h2><MetricsChart :snapshots="metrics.snapshots" field="p99_latency_ms" color="#ff9f68" label="P99 ms" /></article>
      <article class="card"><h2>Database QPS</h2><MetricsChart :snapshots="metrics.snapshots" field="db_qps" color="#6de0b0" label="DB QPS" /></article>
    </section>
    <section class="card read-test"><h2>Read test</h2><div class="inputs"><input v-model.number="entityID" type="number" min="1" max="1000000" /><button @click="fetchFeature">Read feature</button></div><pre>{{ result }}</pre></section>
  </main>
</template>

<style>
:root { color: #e8edf2; background: #10151b; font-family: Inter, system-ui, sans-serif; } body { margin: 0; }
main { max-width: 1180px; margin: 0 auto; padding: 3rem 2rem 5rem; } header { display:flex; justify-content:space-between; align-items:flex-start; gap:2rem; }
.eyebrow { color:#6de0b0; letter-spacing:.14em; font-size:.75rem; } h1 { font-size:clamp(2.2rem, 7vw, 4.8rem); line-height:.95; margin:1rem 0; } h2 { font-size:1rem; margin:0 0 1rem; color:#b9c4cc; }
.intro { color:#9daab5; font-size:1.15rem; } .card { padding:1.25rem; background:#18212a; border:1px solid #293641; border-radius:12px; }
.connection { padding:.35rem .6rem; border:1px solid #7d5d52; color:#ff9f68; border-radius:999px; font-size:.7rem; letter-spacing:.1em; } .connection.online { color:#6de0b0; border-color:#427f68; }
.toolbar { display:flex; align-items:flex-end; gap:2rem; margin:3rem 0 1rem; flex-wrap:wrap; } .label { display:block; color:#a9b5bf; font-size:.8rem; margin-bottom:.5rem; } .segmented, .inputs { display:flex; gap:.4rem; }
button, input { padding:.65rem .8rem; border-radius:7px; border:1px solid #40515e; background:#111820; color:inherit; font:inherit; } button { cursor:pointer; } button.selected, button:hover { background:#6de0b0; color:#102019; } small { color:#71808d; }
.notice { color:#6de0b0; font-size:.85rem; } .stats, .charts { display:grid; grid-template-columns:repeat(4, 1fr); gap:1rem; margin:1rem 0; } .charts { grid-template-columns:repeat(2, 1fr); }
.stat span { color:#91a0ad; font-size:.85rem; } .stat strong { display:block; margin-top:.5rem; font-size:1.6rem; } .read-test { margin-top:1rem; } pre { overflow:auto; color:#b7f4d8; }
@media (max-width: 760px) { main { padding:2rem 1rem; } .stats, .charts { grid-template-columns:1fr; } header { display:block; } .connection { display:inline-block; } }
</style>
