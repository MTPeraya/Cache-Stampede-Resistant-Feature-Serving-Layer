import { onBeforeUnmount, onMounted, ref } from "vue";
import { useMetricsStore } from "./stores/metrics";
import MetricsChart from "./components/MetricsChart.vue";
const metrics = useMetricsStore();
const mode = ref("naive");
const burstRPS = ref(200);
const burstSeconds = ref(10);
const burstKeys = ref(10);
const notice = ref("");
const entityID = ref(1);
const result = ref("No request yet");
async function setMode(next) {
    const response = await fetch(`/mode?value=${next}`, { method: "POST" });
    if (!response.ok) {
        notice.value = "Could not change mode";
        return;
    }
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
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.main, __VLS_intrinsicElements.main)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.header, __VLS_intrinsicElements.header)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "eyebrow" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "intro" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "connection" },
    ...{ class: ({ online: __VLS_ctx.metrics.connected }) },
});
(__VLS_ctx.metrics.connected ? "LIVE" : "OFFLINE");
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "toolbar card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "segmented" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.setMode('naive');
        } },
    ...{ class: ({ selected: __VLS_ctx.mode === 'naive' }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (...[$event]) => {
            __VLS_ctx.setMode('hardened');
        } },
    ...{ class: ({ selected: __VLS_ctx.mode === 'hardened' }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "label" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "inputs" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    type: "number",
    min: "1",
});
(__VLS_ctx.burstRPS);
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    type: "number",
    min: "1",
});
(__VLS_ctx.burstSeconds);
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    type: "number",
    min: "1",
});
(__VLS_ctx.burstKeys);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.runBurst) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "notice" },
});
(__VLS_ctx.notice);
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "stats" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(((__VLS_ctx.metrics.latest?.hit_rate ?? 0) * 100).toFixed(1));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
((__VLS_ctx.metrics.latest?.p99_latency_ms ?? 0).toFixed(1));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
((__VLS_ctx.metrics.latest?.db_qps ?? 0).toFixed(1));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "stat card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.metrics.latest?.coalesced ?? 0);
(__VLS_ctx.metrics.latest?.stale_serves ?? 0);
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "charts" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.article, __VLS_intrinsicElements.article)({
    ...{ class: "card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
/** @type {[typeof MetricsChart, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(MetricsChart, new MetricsChart({
    snapshots: (__VLS_ctx.metrics.snapshots),
    field: "p99_latency_ms",
    color: "#ff9f68",
    label: "P99 ms",
}));
const __VLS_1 = __VLS_0({
    snapshots: (__VLS_ctx.metrics.snapshots),
    field: "p99_latency_ms",
    color: "#ff9f68",
    label: "P99 ms",
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
__VLS_asFunctionalElement(__VLS_intrinsicElements.article, __VLS_intrinsicElements.article)({
    ...{ class: "card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
/** @type {[typeof MetricsChart, ]} */ ;
// @ts-ignore
const __VLS_3 = __VLS_asFunctionalComponent(MetricsChart, new MetricsChart({
    snapshots: (__VLS_ctx.metrics.snapshots),
    field: "db_qps",
    color: "#6de0b0",
    label: "DB QPS",
}));
const __VLS_4 = __VLS_3({
    snapshots: (__VLS_ctx.metrics.snapshots),
    field: "db_qps",
    color: "#6de0b0",
    label: "DB QPS",
}, ...__VLS_functionalComponentArgsRest(__VLS_3));
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "card read-test" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "inputs" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    type: "number",
    min: "1",
    max: "1000000",
});
(__VLS_ctx.entityID);
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ onClick: (__VLS_ctx.fetchFeature) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.pre, __VLS_intrinsicElements.pre)({});
(__VLS_ctx.result);
/** @type {__VLS_StyleScopedClasses['eyebrow']} */ ;
/** @type {__VLS_StyleScopedClasses['intro']} */ ;
/** @type {__VLS_StyleScopedClasses['connection']} */ ;
/** @type {__VLS_StyleScopedClasses['toolbar']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['label']} */ ;
/** @type {__VLS_StyleScopedClasses['segmented']} */ ;
/** @type {__VLS_StyleScopedClasses['label']} */ ;
/** @type {__VLS_StyleScopedClasses['inputs']} */ ;
/** @type {__VLS_StyleScopedClasses['notice']} */ ;
/** @type {__VLS_StyleScopedClasses['stats']} */ ;
/** @type {__VLS_StyleScopedClasses['stat']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['stat']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['charts']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['card']} */ ;
/** @type {__VLS_StyleScopedClasses['read-test']} */ ;
/** @type {__VLS_StyleScopedClasses['inputs']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            MetricsChart: MetricsChart,
            metrics: metrics,
            mode: mode,
            burstRPS: burstRPS,
            burstSeconds: burstSeconds,
            burstKeys: burstKeys,
            notice: notice,
            entityID: entityID,
            result: result,
            setMode: setMode,
            runBurst: runBurst,
            fetchFeature: fetchFeature,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
