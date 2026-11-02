import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import uPlot from "uplot";
import "uplot/dist/uPlot.min.css";
const props = defineProps();
const chartElement = ref();
let chart;
function render() {
    if (!chartElement.value)
        return;
    const data = [
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
    }
    else {
        chart.setData(data);
    }
}
onMounted(render);
watch(() => props.snapshots, render, { deep: true });
onBeforeUnmount(() => chart?.destroy());
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ref: "chartElement",
    ...{ class: "chart" },
});
/** @type {typeof __VLS_ctx.chartElement} */ ;
/** @type {__VLS_StyleScopedClasses['chart']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            chartElement: chartElement,
        };
    },
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
