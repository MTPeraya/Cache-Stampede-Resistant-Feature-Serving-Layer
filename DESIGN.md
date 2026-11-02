# Visual system

## Direction

An Operate-mode observability control room expressed as a Wim Crouwel-inspired type specimen: the visible construction grid is the measurement armature, not decoration. Information is arranged as measured bands and panels, with a single yellow signal plane for active controls and primary metrics.

## Roles

- **Interface voice:** IBM Plex Sans for headings, explanations, and controls.
- **Measurement voice:** IBM Plex Mono for labels, live values, timestamps, and API output.
- **State color:** yellow for selected/actionable/primary signal, mint for healthy live state, muted gray-green for secondary information.

## Structure

- Header establishes stream state and product identity.
- Intro establishes the experiment and active read path.
- Control strip keeps mode and burst dispatch adjacent.
- Metrics band gives the first glance to cache hit rate, then latency and throughput.
- Charts carry the live evidence.
- Chaos and single-request tools sit below the evidence they affect.

## Constraints

- Preserve the existing Vue 3, Pinia, uPlot, SSE, and API behavior.
- Keep keyboard focus visible and controls usable at mobile widths.
- Keep the grid legible as a measurement surface; do not add decorative illustration or fabricated data.
