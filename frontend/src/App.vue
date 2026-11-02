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
  notice.value = response.ok ? `${next} read path selected` : "Mode change failed";
  if (response.ok) mode.value = next;
}

async function runBurst() {
  const response = await fetch(`/loadgen/burst?rps=${burstRPS.value}&seconds=${burstSeconds.value}&keys=${burstKeys.value}`, { method: "POST" });
  notice.value = response.ok ? "Burst dispatched" : "Burst parameters rejected";
}

async function chaos(path: string) {
  const response = await fetch(path, { method: "POST" });
  notice.value = response.ok ? "Failure injection applied" : "Failure injection failed";
}

async function fetchFeature() {
  const response = await fetch(`/features/${entityID.value}`);
  result.value = response.ok ? JSON.stringify(await response.json(), null, 2) : `HTTP ${response.status}`;
}

onMounted(metrics.connect);
onBeforeUnmount(metrics.disconnect);
</script>

<template>
  <main class="shell">
    <header class="topbar">
      <a class="brand" href="/" aria-label="Stampede Shield home"><span class="brand-mark">SS</span><span>STAMPede shield</span></a>
      <div class="top-meta"><span>OBSERVABILITY / 03</span><span class="status"><i :class="{ active: metrics.connected }"></i>{{ metrics.connected ? "STREAM LIVE" : "STREAM OFFLINE" }}</span></div>
    </header>

    <section class="intro">
      <div>
        <p class="section-label">Feature serving / control room</p>
        <h1>See the herd<br /><em>before it hits.</em></h1>
        <p class="lede">Compare cache behavior under pressure. Run a burst, inject a failure, and watch the read path reveal its shape.</p>
      </div>
      <div class="mode-display" aria-label="Active read path">
        <span class="section-label">Active read path</span>
        <strong>{{ mode === "naive" ? "N" : "H" }}</strong>
        <span class="mode-name">{{ mode }}</span>
      </div>
    </section>

    <section class="control-grid" aria-label="Experiment controls">
      <div class="control-block">
        <span class="section-label">01 / read path</span>
        <div class="segmented">
          <button :class="{ selected: mode === 'naive' }" @click="setMode('naive')">Naive</button>
          <button :class="{ selected: mode === 'hardened' }" @click="setMode('hardened')">Hardened</button>
        </div>
      </div>
      <div class="control-block burst-control">
        <span class="section-label">02 / dispatch burst</span>
        <div class="field-row">
          <label><span>RPS</span><input v-model.number="burstRPS" type="number" min="1" /></label>
          <label><span>SEC</span><input v-model.number="burstSeconds" type="number" min="1" /></label>
          <label><span>KEYS</span><input v-model.number="burstKeys" type="number" min="1" /></label>
          <button class="accent-button" @click="runBurst">Run burst <span>↗</span></button>
        </div>
      </div>
      <p class="notice" role="status">{{ notice }}</p>
    </section>

    <section class="metrics-band" aria-label="Live metrics">
      <div class="metric metric-primary"><span class="section-label">Cache hit rate</span><strong>{{ ((metrics.latest?.hit_rate ?? 0) * 100).toFixed(1) }}<small>%</small></strong></div>
      <div class="metric"><span class="section-label">P99 latency</span><strong>{{ (metrics.latest?.p99_latency_ms ?? 0).toFixed(1) }}<small>ms</small></strong></div>
      <div class="metric"><span class="section-label">Database QPS</span><strong>{{ (metrics.latest?.db_qps ?? 0).toFixed(1) }}</strong></div>
      <div class="metric"><span class="section-label">Coalesced / stale</span><strong>{{ metrics.latest?.coalesced ?? 0 }} <small>/</small> {{ metrics.latest?.stale_serves ?? 0 }}</strong></div>
      <div class="metric"><span class="section-label">In flight / shed</span><strong>{{ metrics.latest?.in_flight_db ?? 0 }} <small>/</small> {{ metrics.latest?.shed_requests ?? 0 }}</strong></div>
    </section>

    <section class="chart-grid" aria-label="Live charts">
      <article class="chart-panel">
        <div class="panel-heading"><div><span class="section-label">Latency / rolling 2 min</span><h2>P99 request latency</h2></div><span class="chart-value">{{ (metrics.latest?.p99_latency_ms ?? 0).toFixed(1) }} ms</span></div>
        <MetricsChart :snapshots="metrics.snapshots" field="p99_latency_ms" color="#f4b942" label="P99 ms" />
      </article>
      <article class="chart-panel">
        <div class="panel-heading"><div><span class="section-label">Throughput / rolling 2 min</span><h2>Database QPS</h2></div><span class="chart-value green">{{ (metrics.latest?.db_qps ?? 0).toFixed(1) }} qps</span></div>
        <MetricsChart :snapshots="metrics.snapshots" field="db_qps" color="#7ee2b8" label="DB QPS" />
      </article>
    </section>

    <section class="lower-grid">
      <div class="failure-panel">
        <div class="panel-heading"><div><span class="section-label">03 / controlled failure</span><h2>Chaos panel</h2></div><span class="signal">INJECT</span></div>
        <div class="failure-actions">
          <button @click="chaos('/chaos/flush-cache')">Flush cache <span>↗</span></button>
          <button @click="chaos('/chaos/redis-down?seconds=15')">Redis down / 15s <span>↗</span></button>
          <button @click="chaos('/chaos/db-latency?ms=250')">DB latency / 250ms <span>↗</span></button>
          <button @click="chaos('/chaos/db-latency?ms=0')">Clear latency <span>↗</span></button>
        </div>
      </div>
      <div class="read-panel">
        <div class="panel-heading"><div><span class="section-label">04 / single request</span><h2>Read test</h2></div></div>
        <div class="field-row"><label><span>ENTITY ID</span><input v-model.number="entityID" type="number" min="1" max="1000000" /></label><button class="accent-button" @click="fetchFeature">Read feature <span>↗</span></button></div>
        <pre>{{ result }}</pre>
      </div>
    </section>
  </main>
</template>

<style>
@import url("https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500&family=IBM+Plex+Sans:wght@400;500;600;700&display=swap");

:root {
  color: #e8e4da;
  background: #101311;
  font-family: "IBM Plex Sans", system-ui, sans-serif;
  font-synthesis: none;
  text-rendering: optimizeLegibility;
}
* { box-sizing: border-box; }
body { margin: 0; min-width: 320px; background: #101311; }
button, input { font: inherit; }
button { cursor: pointer; }
button:focus-visible, input:focus-visible, a:focus-visible { outline: 2px solid #f4b942; outline-offset: 3px; }
::selection { color: #101311; background: #f4b942; }
.shell { --line: rgba(232, 228, 218, .16); --muted: #8d958d; --dim: #59615a; --green: #7ee2b8; --yellow: #f4b942; max-width: 1440px; margin: 0 auto; padding: 0 42px 72px; background-image: linear-gradient(rgba(232,228,218,.045) 1px, transparent 1px), linear-gradient(90deg, rgba(232,228,218,.045) 1px, transparent 1px); background-size: 32px 32px; }
.topbar { height: 74px; border-bottom: 1px solid var(--line); display:flex; align-items:center; justify-content:space-between; font-family:"IBM Plex Mono", monospace; font-size:11px; letter-spacing:.08em; text-transform:uppercase; }
.brand, .top-meta, .status { display:flex; align-items:center; gap:12px; color:#e8e4da; text-decoration:none; } .brand-mark { display:grid; place-items:center; width:27px; height:27px; color:#101311; background:var(--yellow); font-weight:500; letter-spacing:-.08em; } .top-meta { color:var(--muted); gap:28px; } .status { color:var(--green); } .status i { width:7px; height:7px; border-radius:50%; background:#805f36; } .status i.active { background:var(--green); box-shadow:0 0 0 4px rgba(126,226,184,.12); }
.intro { min-height:340px; display:grid; grid-template-columns: minmax(0, 1fr) 260px; gap:40px; align-items:end; padding:72px 0 58px; border-bottom:1px solid var(--line); }
.section-label { display:block; color:var(--muted); font:500 10px/1.2 "IBM Plex Mono", monospace; letter-spacing:.12em; text-transform:uppercase; } h1 { max-width:720px; margin:16px 0 18px; font-size:clamp(48px, 7vw, 94px); line-height:.92; letter-spacing:-.06em; font-weight:600; text-wrap:balance; } h1 em { color:var(--yellow); font-style:normal; } .lede { max-width:55ch; margin:0; color:#aab0a8; font-size:16px; line-height:1.55; text-wrap:pretty; }
.mode-display { align-self:stretch; display:flex; flex-direction:column; justify-content:space-between; padding:18px 0 0 24px; border-left:1px solid var(--line); } .mode-display strong { margin-top:auto; color:var(--yellow); font:400 160px/.78 "IBM Plex Mono", monospace; letter-spacing:-.12em; } .mode-name { color:#e8e4da; font:500 13px "IBM Plex Mono", monospace; letter-spacing:.1em; text-transform:uppercase; }
.control-grid { display:grid; grid-template-columns:220px minmax(0,1fr) auto; gap:34px; align-items:end; padding:26px 0; border-bottom:1px solid var(--line); } .control-block { min-height:70px; } .segmented, .field-row { display:flex; gap:8px; align-items:end; margin-top:12px; } .segmented button, .failure-actions button { padding:9px 14px; color:var(--muted); border:1px solid var(--line); background:rgba(16,19,17,.5); font-size:12px; } .segmented button.selected { color:#101311; border-color:var(--yellow); background:var(--yellow); } .field-row label { display:grid; gap:5px; color:var(--dim); font:10px "IBM Plex Mono", monospace; letter-spacing:.08em; } input { width:72px; padding:9px 10px; color:#e8e4da; border:1px solid var(--line); background:#171b17; font:12px "IBM Plex Mono", monospace; } .accent-button { padding:10px 15px; color:#101311; border:1px solid var(--yellow); background:var(--yellow); font-size:12px; font-weight:600; } button span { margin-left:12px; } .notice { margin:0; color:var(--green); font:11px "IBM Plex Mono", monospace; min-width:150px; text-align:right; }
.metrics-band { display:grid; grid-template-columns:1.3fr repeat(4,1fr); border-bottom:1px solid var(--line); } .metric { min-height:130px; padding:24px 18px 18px; border-left:1px solid var(--line); } .metric:first-child { border-left:0; } .metric-primary { background:var(--yellow); color:#101311; } .metric-primary .section-label { color:#514112; } .metric strong { display:block; margin-top:22px; color:#e8e4da; font:400 28px "IBM Plex Mono", monospace; letter-spacing:-.08em; font-variant-numeric:tabular-nums; } .metric-primary strong { color:#101311; font-size:42px; } .metric small { font-size:12px; letter-spacing:0; }
.chart-grid { display:grid; grid-template-columns:1fr 1fr; gap:1px; background:var(--line); border-bottom:1px solid var(--line); } .chart-panel { min-width:0; min-height:318px; padding:26px 24px 14px; background:#101311; } .panel-heading { display:flex; justify-content:space-between; align-items:start; gap:16px; } h2 { margin:7px 0 0; font-size:20px; letter-spacing:-.03em; font-weight:500; } .chart-value { color:var(--yellow); font:12px "IBM Plex Mono", monospace; } .chart-value.green { color:var(--green); }
.lower-grid { display:grid; grid-template-columns:1fr 1fr; gap:1px; background:var(--line); } .failure-panel, .read-panel { min-width:0; min-height:220px; padding:26px 24px; background:#141815; } .read-panel { background:#171b17; } .signal { color:var(--yellow); font:10px "IBM Plex Mono", monospace; letter-spacing:.1em; } .failure-actions { display:grid; grid-template-columns:1fr 1fr; gap:8px; margin-top:28px; } .failure-actions button { display:flex; justify-content:space-between; text-align:left; } .failure-actions button:hover { color:#101311; border-color:var(--yellow); background:var(--yellow); } .read-panel .field-row { margin-top:28px; } pre { min-height:52px; max-height:80px; overflow:auto; margin:24px 0 0; padding-top:14px; border-top:1px solid var(--line); color:var(--green); font:11px/1.5 "IBM Plex Mono", monospace; }
@media (max-width: 900px) { .shell { padding:0 24px 52px; } .control-grid { grid-template-columns:1fr 1fr; } .notice { grid-column:1 / -1; text-align:left; } .metrics-band { grid-template-columns:repeat(3,1fr); } .metric-primary { grid-row:span 1; } .metric:nth-child(4), .metric:nth-child(5) { border-top:1px solid var(--line); } }
@media (max-width: 620px) { .shell { padding:0 16px 40px; background-size:24px 24px; } .topbar { height:62px; } .top-meta > span:first-child { display:none; } .intro { min-height:0; grid-template-columns:1fr; gap:32px; padding:52px 0 38px; } h1 { font-size:54px; } .mode-display { min-height:120px; padding:16px 0 0; border-left:0; border-top:1px solid var(--line); } .mode-display strong { font-size:90px; margin-top:34px; } .control-grid, .chart-grid, .lower-grid { grid-template-columns:1fr; } .control-grid { gap:26px; } .burst-control .field-row, .read-panel .field-row { flex-wrap:wrap; } .metrics-band { grid-template-columns:1fr 1fr; } .metric { min-height:104px; padding:18px 12px 14px; } .metric strong, .metric-primary strong { margin-top:14px; font-size:24px; } .metric-primary { grid-column:span 2; } .metric:nth-child(3) { border-left:0; } .chart-panel { min-height:270px; padding:22px 14px 10px; } .failure-actions { grid-template-columns:1fr; } }
</style>
