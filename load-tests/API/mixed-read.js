import http from "k6/http";
import { check, sleep } from "k6";

const BASE_URL = "http://localhost:8080";

export const options = {
  scenarios: {
    health: {
      executor: "constant-arrival-rate",
      rate: 500,
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 200,
      maxVUs: 1000,
      exec: "health",
    },
    projects: {
      executor: "constant-arrival-rate",
      rate: 300,
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 200,
      maxVUs: 1000,
      exec: "projects",
    },
    app_url: {
      executor: "constant-arrival-rate",
      rate: 100,
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 100,
      maxVUs: 500,
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
  const res = http.get(`${BASE_URL}/projects/14/url`);
  check(res, { "url 200": (r) => r.status === 200 });
}

export default function () {
  sleep(1);
}