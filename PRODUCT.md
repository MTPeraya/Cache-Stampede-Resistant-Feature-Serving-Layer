# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

A software engineer or reviewer evaluating cache-stampede behavior during a live demo.

## Product Purpose

Make naive versus hardened feature serving behavior immediately understandable through live metrics and controlled experiments.

## Capabilities and Constraints

- Vue 3, Pinia, and uPlot frontend.
- Live metrics arrive through Server-Sent Events.
- Preserve existing dashboard controls and API behavior.
- Responsive desktop and mobile use.
- Use truthful measured data only; do not fabricate benchmark results.
- The dashboard is an Operate-mode surface for running experiments and reading live system state.

## Evidence on Hand

- Live metrics dashboard in `frontend/src/App.vue`.
- Pinia SSE consumer in `frontend/src/stores/metrics.ts`.
- uPlot chart component in `frontend/src/components/MetricsChart.vue`.
- Backend APIs and chaos controls under `backend/internal/api`.
- No demo GIF or measured benchmark dataset is currently available.

## Product Principles

- Make system behavior legible at a glance.
- Keep experiment controls close to the evidence they affect.
- Prefer honest live state over decorative simulation.
- Preserve familiar dashboard interaction patterns.
