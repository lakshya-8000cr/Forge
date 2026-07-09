import http from "k6/http";
import { check, sleep } from "k6";

const BASE_URL = "http://localhost:8080";
const PROJECT_ID = 7;

// MERGED OPTIONS: Ek hi block mein sab lock kar diya hai
export const options = {
  noConnectionReuse: false, // Keep-Alive hot connections enabled! 🏎️
  userAgent: "k6-forge-load-test",
  
  scenarios: {
    health: {
      executor: "constant-arrival-rate",
      rate: 500,
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 10000,
      maxVUs: 10000,
      exec: "health",
    },
    projects: {
      executor: "constant-arrival-rate",
      rate: 300,
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 10000,
      maxVUs: 10000, // FIXED: Bounded properly (Must be >= preAllocatedVUs)
      exec: "projects",
    },
    app_url: {
      executor: "constant-arrival-rate",
      rate: 100,
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 1000,
      maxVUs: 1000, // FIXED: Aligned seamlessly
      exec: "appUrl",
    },
  },
};

export function health() {
  const res = http.get(`${BASE_URL}/health`);
  check(res, { "health 200": (r) => r.status === 200 });
}

export function projects() {
  const res = http.get(`${BASE_URL}/projects`);
  check(res, { "projects 200": (r) => r.status === 200 });
}

export function appUrl() {
  const res = http.get(`${BASE_URL}/projects/${PROJECT_ID}/url`);
  check(res, { "url 200": (r) => r.status === 200 });
}

export default function () {
  sleep(1);
}