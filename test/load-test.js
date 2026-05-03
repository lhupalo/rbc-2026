import http from "k6/http";
import { check } from "k6";
import { Counter } from "k6/metrics";

const payloads = JSON.parse(open("./test-data.json"));
const baseURL = __ENV.BASE_URL || "http://localhost:9999";
const endpoint = `${baseURL}/fraud-score`;

export const options = {
  scenarios: {
    ramp_up: {
      executor: "ramping-vus",
      startVUs: 1,
      stages: [
        { duration: "20s", target: 20 },
        { duration: "20s", target: 50 },
        { duration: "30s", target: 100 },
        { duration: "20s", target: 0 },
      ],
      gracefulRampDown: "5s",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(99)<2000"],
  },
};

const responseSchemaErrors = new Counter("response_schema_errors");

export default function () {
  const payload = payloads[__ITER % payloads.length];
  const response = http.post(endpoint, JSON.stringify(payload), {
    headers: {
      "Content-Type": "application/json",
    },
    timeout: "2000ms",
  });

  const ok = check(response, {
    "status 200": (r) => r.status === 200,
    "approved boolean": (r) => {
      if (r.status !== 200) return false;
      const body = r.json();
      return typeof body.approved === "boolean";
    },
    "fraud_score number": (r) => {
      if (r.status !== 200) return false;
      const body = r.json();
      return typeof body.fraud_score === "number";
    },
  });

  if (!ok) {
    responseSchemaErrors.add(1);
  }
}

export function handleSummary(data) {
  return {
    "results.json": JSON.stringify(data, null, 2),
    stdout: `\nTeste finalizado. p(99): ${data.metrics.http_req_duration.values["p(99)"]} ms\n`,
  };
}
