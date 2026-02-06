This project is a Project and Task Management Application built to help teams and its people organize their work and communicate in one go. The main idea is to make it easier to manage projects, track tasks along different stages, and allow team and its members to collaborate without using a different application.

We are developing this project as part of our Software Engineering course, focusing on real-world collaboration problems and full-stack development.

🎯 Why did We Built This ???

While working in teams, we often notice that tracking down our tasks and communication with members happen on different platforms.This project combines both into a single application, making teamwork simpler and more organized.The goals of this project are:

To manage projects and tasks in a structured manner.To visualize task progress clearly.To enable real-time communication within the team.To apply software engineering concepts in a practical project implementation.Task Stages

Every project contains stages such as Planned, In Progress, and Completed

These Stages help track the progress of tasks

Users can add or remove stages as needed


Task Management

Tasks are created inside stages

Each task has a title and description

Tasks can be edited, moved between stages, or deleted

All task data is stored securely in the database

Integrated Chat System

Each project has its own chat section

Team members can send messages in real time

Messages show who sent them and when

This helps teams communicate without leaving the app

🧠 How the System Works

The application follows a simple client–server architecture.

The Angular frontend handles the user interface

The Go backend processes requests and manages data

SQLite is used to store projects, tasks, and chat messages

The chat system uses WebSockets or API-based communication for real-time updates

Angular Frontend  ↔  Go Backend  ↔  SQLite Database

🛠️ Technologies Used
Frontend

Angular

TypeScript

HTML and CSS

Component-based design

Angular services for API and chat communication

Backend

Go (Golang)

REST APIs

Real-time chat support

Modular and clean code structure

Database

SQLite (relational database)

⚙️ How to Run the Project
Backend
cd backend
go mod tidy
go run main.go

Frontend
cd frontend
npm install
ng serve


The application will be available at:

http://localhost:4200

📌 Future Improvements

User login and authentication

Role-based access for team members

Notifications for task updates

File sharing in chat

Better UI and mobile responsiveness

📄 Conclusion

This project helped us understand how real-world project management systems work and how frontend and backend components communicate with each other. It also gave us hands-on experience with Angular, Go, and real-time features like chat systems.

👥 Team Members & Contributions

This project was developed as a team effort.

Adithya – Backend development, API design, database schema

Nandhan – Backend development, chat system implementation

Meghana – Frontend development, project and task UI

Srija – Frontend development, chat interface and integration