import http from "k6/http";
import { check, sleep } from "k6";

const BASE_URL = "http://a6228a9f5a97041acb1f84aa6ada2478-774918050.eu-north-1.elb.amazonaws.com";

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
  const res = http.get(`${BASE_URL}/api/health`);
  check(res, { "health 200": (r) => r.status === 200 });
}

export function projects() {
  const res = http.get(`${BASE_URL}/api/projects`);
  check(res, { "projects 200": (r) => r.status === 200 });
}

export function appUrl() {
  const res = http.get(`${BASE_URL}/api/projects/14/url`);
  check(res, { "url 200": (r) => r.status === 200 });
}

export default function () {
  sleep(1);
}