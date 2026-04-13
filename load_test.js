// Import necessary libraries.
import http from 'k6/http';
import { check, sleep } from 'k6';

// Options define the load test behavior, simulation stages, and thresholds.
export const options = {
    // 'stages' simulates a realistic spike in network traffic over 1 minute.
    stages: [
        { duration: '15s', target: 5000 }, // Ramp-up: Reach 5000 concurrent users in 15 seconds.
        { duration: '30s', target: 2000 }, // Peak: Hold at 2000 users for 30 seconds.
        { duration: '15s', target: 0 }, // Ramp-down: Gradually drop to 0 user in 15 seconds.
    ],
    // Thresholds match the definition of done (DoD) for performance.
    thresholds: {
        // http_req_failed represents the error rate (status codes like 500, 404, or timeouts).
        http_req_failed:    ['rate < 0.01'], // Ensure the error rate is strictly < 1%.
        // http_red_duration is the response time. p(95) means 95% of requests must be faster than this limit.
        http_req_duration:  ['p(95) < 200'], // P95 response time < 200 ms.
    },
};

// setup() is executed ONCE before the load test starts.
// We use it to authenticate and retrieve a valid JWT token for the virtual users.
export function setup() {
    const loginUrl = 'http://localhost:8080/api/v1/auth/login';
    // Using integer for tenant_id matching the new uint Domain struct.
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
    // Extract and return the token so it can be passed to the default function.
    // If login fails, this will return undefined and subsequent requests will return 401.
    return {
        token: res.json('token')
    };
}

// The default exported function represents the behavior of each virtual user.
// 'data' contains the returned object from the setup() function.
export default function(data) {
    // The target API endpoint (fetching question ID 1, leveraging Redis cache).
    const url = 'http://localhost:8080/api/v1/questions/1';
    // Inject the JWT token securely into the authorization header.
    const params = {
        headers: {
            'Authorization': `Bearer ${data.token}`,
        },
    };
    // Execute the GET request.
    const res = http.get(url, params);
    // Verify that the server returned a 200 OK status.
    check(res, {
        'is status 200': (r) => r.status === 200,
    });
    // Sleep for 1 second to simulate a real human user reading the UI.
    // This prevents a local testing machine from freezing due to infinite loops.
    sleep(1);
}