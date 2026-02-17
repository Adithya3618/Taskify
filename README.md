# 🚀 Project & Task Management Application

A modern **Project and Task Management Application** designed to help teams organize work, track progress, and communicate seamlessly — all within a single platform.

This application brings together **task tracking and real-time collaboration**, eliminating the need to switch between multiple tools. Built as part of our Software Engineering coursework, the project focuses on solving real-world team coordination challenges through a full-stack  implementation.

---

# 🎯 Project Motivation

In many team environments, **task tracking and communication happen across different platforms**, leading to confusion, missed updates, and reduced productivity.

We built this application to bring everything into one unified workspace.

## Key Objectives
- Manage projects and tasks in a structured and organized way  
- Visualize task progress clearly across stages  
- Enable real-time team communication within each project  
- Apply software engineering concepts in a practical full-stack implementation  

---

# ✨ Core Features

## 📋 Task Stages & Workflow
Each project is organized into customizable stages:
- Planned  
- In Progress  
- Completed  

These stages help teams monitor task progress visually.  
Users can also create or remove stages based on project requirements.

---

## 📝 Task Management
Tasks are managed within stages and include:
- Title and description  
- Edit and delete functionality  
- Ability to move tasks between stages  
- Secure storage in the database  

This structured approach improves task visibility, accountability, and workflow tracking.

---

## 💬 Integrated Real-Time Chat
Each project includes its own chat workspace for collaboration.

**Features include:**
- Real-time messaging between team members  
- Sender name and timestamp display  
- Communication without leaving the application  

This ensures all project discussions remain organized and centralized.

---

# 🧠 System Architecture

The application follows a clean client–server architecture:


### How it works
- The Angular frontend handles user interface and interactions  
- The Go backend processes requests and manages application logic  
- SQLite securely stores project, task, and chat data  
- Chat functionality uses WebSockets or API-based real-time communication  

---

# 🛠️ Technology Stack

## Frontend
- Angular  
- TypeScript  
- HTML and CSS  
- Component-based architecture  
- Angular services for API and chat communication  

## Backend
- Go (Golang)  
- RESTful APIs  
- Real-time chat support  
- Modular and clean code structure  

## Database
- SQLite (Relational Database)

---

# ⚙️ Running the Project Locally

## Backend Setup
```bash
cd backend
go mod tidy
go run main.go
```
## Frontend Setup
```bash
cd frontend
npm install
ng serve
```
## Application is at 
```bash
http://localhost:4200
```

# 🔮 Future Enhancements
- User login and authentication system  
- Role-based access control for teams  
- Notifications for task updates  
- File sharing within project chat  
- Improved UI and mobile responsiveness  

---

# 📚 Learning Outcomes
This project provided practical experience in:
- Full-stack application development  
- REST API design and integration  
- Real-time communication systems  
- Frontend and backend integration  
- Building scalable and modular applications  

It also helped us understand how modern project management tools function in real-world environments.

---

# 👥 Team Contributions
- **Adithya** — Backend development 
- **Nandhan** — Backend development  
- **Meghana** — Frontend development  
- **Sai Sreeja** — Frontend development

---

# 📄 Conclusion
This Project & Task Management Application demonstrates how a unified platform can improve team productivity, organization, and collaboration. By combining task tracking and communication into a single system, the project reflects real-world software engineering practices and modern full-stack development.

