import http from "k6/http";
import { check } from "k6";

const BASE_URL = "http://localhost:8080";

// Change this according to your real project IDs
const PROJECT_IDS = [1, 2, 3, 4, 5];

export const options = {
  scenarios: {
    deploy_stress: {
      executor: "ramping-arrival-rate",
      startRate: 1,
      timeUnit: "1s",
      preAllocatedVUs: 200,
      maxVUs: 1000,
      stages: [
        { duration: "30s", target: 10 },
        { duration: "30s", target: 25 },
        { duration: "30s", target: 50 },
        { duration: "30s", target: 100 },
        { duration: "10s", target: 0 },
      ],
    },
  },

  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<200"],
  },
};

export default function () {
  const projectId = PROJECT_IDS[Math.floor(Math.random() * PROJECT_IDS.length)];

  const url = `${BASE_URL}/projects/${projectId}/deploy`;

  const res = http.post(url, null, {
    headers: {
      "Content-Type": "application/json",
    },
  });

  check(res, {
    "202 queued or 409 already deploying": (r) =>
      r.status === 202 || r.status === 409,

    "response body valid": (r) =>
      r.body &&
      (r.body.includes("deployment queued") ||
        r.body.includes("already in progress")),
  });
}