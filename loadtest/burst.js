import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";

const errors = new Counter("feature_errors");
export const options = {
  scenarios: {
    baseline_then_spike: {
      executor: "ramping-arrival-rate",
      startRate: 20,
      timeUnit: "1s",
      preAllocatedVUs: 20,
      maxVUs: 300,
      stages: [
        { target: 20, duration: "30s" },
        { target: 200, duration: "15s" },
        { target: 20, duration: "30s" },
      ],
    },
  },
  thresholds: { http_req_failed: ["rate<0.05"] },
};

export default function () {
  const response = http.get(`http://${__ENV.API_HOST || "localhost:8080"}/features/${(__VU % 1000000) + 1}`);
  if (!check(response, { "feature request succeeds": (r) => r.status === 200 })) errors.add(1);
  sleep(0.01);
}
