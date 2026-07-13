import http from "k6/http";
import { check } from "k6";

const BASE_URL =
  __ENV.BASE_URL ||
  "http://localhost:8080";

const PROJECT_ID = __ENV.PROJECT_ID || "7";

export const options = {
  discardResponseBodies: true,
  noConnectionReuse: false,
  userAgent: "k6-forge-load-test",

  scenarios: {
    health: {
      executor: "constant-arrival-rate",
      rate: 4000,
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 300,
      maxVUs: 2000,
      exec: "health",
      tags: { endpoint: "health" },
    },

    projects: {
      executor: "constant-arrival-rate",
      rate: 3500,
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 500,
      maxVUs: 3000,
      exec: "projects",
      tags: { endpoint: "projects" },
    },

    app_url: {
      executor: "constant-arrival-rate",
      rate: 2500,
      timeUnit: "1s",
      duration: "2m",
      preAllocatedVUs: 400,
      maxVUs: 2500,
      exec: "appUrl",
      tags: { endpoint: "app_url" },
    },
  },

  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<100", "p(99)<250"],

    "http_req_failed{endpoint:health}": ["rate<0.01"],
    "http_req_failed{endpoint:projects}": ["rate<0.01"],
    "http_req_failed{endpoint:app_url}": ["rate<0.01"],
  },
};

const params = {
  headers: {
    Connection: "keep-alive",
    Accept: "application/json",
  },
};

export function health() {
  const res = http.get(`${BASE_URL}/health`, params);

  check(res, {
    "health returned 200": (r) => r.status === 200,
  });
}

export function projects() {
  const res = http.get(`${BASE_URL}/projects`, params);

  check(res, {
    "projects returned 200": (r) => r.status === 200,
  });
}

export function appUrl() {
  const res = http.get(
    `${BASE_URL}/projects/${PROJECT_ID}/url`,
    params,
  );

  check(res, {
    "url returned 200": (r) => r.status === 200,
  });
}