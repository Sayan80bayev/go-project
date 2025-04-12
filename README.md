# go-project

![Go Version](https://img.shields.io/badge/Go-1.21+-brightgreen)
![Gin Framework](https://img.shields.io/badge/Gin-Framework-blue)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Build](https://img.shields.io/badge/build-passing-brightgreen)

A simple modular Go backend project built with [Gin](https://github.com/gin-gonic/gin).  
This project is educational and suitable for students learning REST API design and Go web development.  
It includes services for users, posts, authentication, and (soon) likes.

---

## 🚀 Features

- ✨ Fast and lightweight Gin web framework
- 🔐 Auth system (JWT-based)
- 📝 Post creation and management
- 👤 User registration and profile handling
- ❤️ Like service (Coming soon!)
- 📦 Modular architecture (per service)
- 🧪 Ideal for students and educational use

---

## 📁 Services Overview

| Service       | Description                         | Path/Module               |
|---------------|-------------------------------------|---------------------------|
| `authService` | Handles login, registration, JWT    | `/services/authService`   |
| `userService` | Manages users and profiles          | `/services/userService`   |
| `postService` | Create/read posts, manage feed      | `/services/postService`   |
| `likeService` | (Planned) Like and unlike posts     | `/services/likeSerevice`  |

---

## 📦 Installation

```bash
git clone https://github.com/Sayan80bayev/go-project.git
cd go-project
docker-compose up
```

📚 Requirements:
  - PostgreSQL
  - Docker
  - Go 1.21+
  - Gin

🧠 For Students

This project was built with clarity and structure in mind — ideal for practicing:
	•	REST API development
	•	Authentication flows
	•	Modular service separation
 
🧱 Folder Structure
```bash
.
├── services/
│   ├── authService/
│   ├── userService/
│   ├── postService/
│   └── likeService/           # Coming soon
└── docker-compose.yaml
```
📄 License

MIT License © 2025
