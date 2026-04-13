// Import necessary libraries.
import http from 'k6/http';
import { check, sleep } from 'k6';

// Options define the load test behavior, simulation stages, and thresholds.
export const options = {
    // 'stages' simulates a realistic and smooth spike in network traffic.
    stages: [
        { duration: '15s', target: 2000 }, // Ramp-up: Smoothly reach 2000 users in order to avoid instant cliff drops.
        { duration: '30s', target: 2000 }, // Peak: Hold at 2000 concurrent virtual users.
        { duration: '15s', target: 0 }, // Ramp-down: Gradually drain active connections.
    ],
    // Thresholds match the definition of done (DoD) for performance.
    thresholds: {
        // http_req_failed represents the error rate (status codes like 500, 404, or timeouts).
        http_req_failed:    ['rate < 0.01'], // Must be strictly < 1% failure.
        // http_red_duration is the response time limit.
        http_req_duration:  ['p(95) < 200'], // 95% of responses must be faster than 200 ms.
    },
};

// setup() is executed ONCE before the load test starts.
// Used to authenticate and retrieve a valid JWT token.
export function setup() {
    const loginUrl = 'http://localhost:8080/api/v1/auth/login';
    // Payload uses standard JSON matching the Go struct.
    const payload = JSON.stringify({
        tenant_id:  1,
        username:   'admin', // Replace with a valid username from your database.
        password:   'password' // Replace with a valid password from your database.
    });
    const params = {
        headers: {
            'Content-Type': 'application/json'
        },
    };
    const res = http.post(loginUrl, payload, params);
    // Validate setup succeeded to avoid cascading failures.
    if (res.status !== 200) {
        console.error(`Login failed! Status: ${res.status}, Body: ${res.body}`);
    }
    return {
        token: res.json('token')
    };
}

// The default function represents the exact behavior executed by each virtual user concurrently.
export default function(data) {
    // Abort early if setup failed to get a token.
    if (!data || !data.token) return;
    // The target API endpoint (fetching question ID 1 to test Redis cache and MySQL fallback performance).
    const url = 'http://localhost:8080/api/v1/questions/1';
    // Inject the JWT token securely.
    const params = {
        headers: {
            'Authorization': `Bearer ${data.token}`,
        },
    };
    const res = http.get(url, params);
    // Verify success.
    check(res, {
        'is status 200': (r) => r.status === 200,
    });
    // Pace the user to prevent machine overload (simulating reading time).
    sleep(1);
}