# Assessment Core Engine API 🚀

A high-performance, multi-tenant Assessment and Online Examination Core Engine built with Go and Fiber. This backend system is designed to handle thousands of concurrent students, real-time proctoring, and complex grading algorithms safely and securely.

## ✨ Key Features

* **Multi-Tenant Architecture:** Securely isolate data (Exams, Users, Questions) across different schools or applications.
* **Advanced RBAC (Role-Based Access Control):** Powered by [Casbin](https://casbin.org/), featuring granular permissions for `SuperAdmin`, `TenantAdmin`, `Teacher`, and `Student`.
* **Proctoring & Anti-Cheat System:** Real-time tracking of student behaviors during exams (e.g., tab switching, exiting full-screen).
* **Password Recovery & OTP:** Secure 6-digit OTP generation via Redis and email delivery using `crypto/rand` for maximum security.
* **Cloud Media Storage:** Direct integration with AWS S3 for uploading and serving exam media (Listening audio, images).
* **Redis Caching & Auto-Save:** High-speed caching for exam fetching and draft auto-saving to prevent data loss.
* **Swagger API Documentation (`fiber-swagger`):** Automatically generates beautiful, interactive API docs for seamless frontend integration and testing.

## 🛠 Tech Stack

* **Language:** [Go (Golang)](https://golang.org/)
* **Web Framework:** [Fiber v2](https://gofiber.io/)
* **Database & ORM:** MySQL 8.0 + [GORM](https://gorm.io/)
* **Caching & In-Memory DB:** [Redis](https://redis.io/)
* **Authentication & Security:** JWT (JSON Web Tokens), bcrypt, Casbin
* **Cloud Storage:** AWS S3 SDK (v2)
* **Validation:** go-playground/validator

## 📖 Swagger API Documentation

This project utilizes `fiber-swagger` to serve auto-generated, interactive API documentation. 

To access the Swagger UI:
1. Start the server.
2. Navigate to: `http://localhost:8080/swagger/` in your browser.

*From here, frontend developers can explore all available endpoints, required payloads, and test the APIs directly without needing Postman!*

## 🚀 Getting Started

### Prerequisites
* Go 1.20 or higher
* MySQL Database
* Redis Server

### Installation

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/tuan-hoang-le/training-2026-fe-be-intern.git](https://github.com/tuan-hoang-le/training-2026-fe-be-intern.git)
   cd training-2026-fe-be-intern