import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
export default defineConfig({
  plugins: [vue()],
  server: { proxy: { "/features": "http://api:8080", "/mode": "http://api:8080", "/metrics": "http://api:8080", "/loadgen": "http://api:8080" } },
});
