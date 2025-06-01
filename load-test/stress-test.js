import http from "k6/http";
import { check } from "k6";

export const options = {
  stages: [
    { duration: "1m", target: 100 },
    { duration: "3m", target: 500 },
    { duration: "2m", target: 1000 },
    { duration: "2m", target: 1000 },
    { duration: "1m", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<1000"],
    http_req_failed: ["rate<0.05"],
  },
};

export default function () {
  const response = http.post(
    "http://host.docker.internal:8080/api/v1/vote",
    JSON.stringify({
      participanteId: Math.floor(Math.random() * 10) + 1,
      sessionId: `stress-${__VU}-${__ITER}`,
    }),
    { headers: { "Content-Type": "application/json" } }
  );

  check(response, {
    "status é 202": (r) => r.status === 202,
  });
}
