import http from "k6/http";
import { sleep, check } from "k6";

export const options = {
  vus: 3000,
  duration: "2m",
};

export default function () {
  const BASE_URL = "http://a6228a9f5a97041acb1f84aa6ada2478-774918050.eu-north-1.elb.amazonaws.com";
  
  const res = http.get(`${BASE_URL}/apps/whoami-demo`);

  check(res, {
    "status is 200": (r) => r.status === 200,
  });

  sleep(1);
}