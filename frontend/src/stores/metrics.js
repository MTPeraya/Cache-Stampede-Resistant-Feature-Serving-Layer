import { defineStore } from "pinia";
import { computed, ref } from "vue";
export const useMetricsStore = defineStore("metrics", () => {
    const snapshots = ref([]);
    const connected = ref(false);
    let source;
    const latest = computed(() => snapshots.value[snapshots.value.length - 1]);
    function connect() {
        source?.close();
        source = new EventSource("/metrics/stream");
        source.onopen = () => { connected.value = true; };
        source.onerror = () => { connected.value = false; };
        source.onmessage = (event) => {
            const snapshot = JSON.parse(event.data);
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
