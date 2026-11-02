import { defineStore } from "pinia";
import { computed, ref } from "vue";

export interface MetricsSnapshot {
  timestamp: string;
  requests: number;
  cache_hits: number;
  cache_misses: number;
  coalesced: number;
  stale_serves: number;
  db_queries: number;
  db_qps: number;
  hit_rate: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
  in_flight_db: number;
  shed_requests: number;
  mode: string;
  chaos_events: string[];
}

export const useMetricsStore = defineStore("metrics", () => {
  const snapshots = ref<MetricsSnapshot[]>([]);
  const connected = ref(false);
  let source: EventSource | undefined;

  const latest = computed(() => snapshots.value[snapshots.value.length - 1]);
  function connect() {
    source?.close();
    source = new EventSource("/metrics/stream");
    source.onopen = () => { connected.value = true; };
    source.onerror = () => { connected.value = false; };
    source.onmessage = (event) => {
      const snapshot = JSON.parse(event.data) as MetricsSnapshot;
      snapshots.value = [...snapshots.value, snapshot].slice(-240);
    };
  }
  function disconnect() {
    source?.close();
    source = undefined;
    connected.value = false;
  }
  return { snapshots, latest, connected, connect, disconnect };
});
