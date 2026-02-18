# Sprint 1 : Taskify

## User Stories

**User Story 1: Project Setup, Dependencies, and Database**

As a developer, To enable the team to begin developing features right away, I want the frontend and backend projects started with the proper folder structure, all necessary packages loaded, and a functional database.


Related Github Issues are : 24, 25, 28

---

**User Story 2: Homepage, Navigation, and Responsive Layout**
As a user, When something goes wrong, I want the program to provide me with clear error messages and the relevant status codes so I can figure out what went wrong and fix it.

Related Github Issues are : 5, 26, 27, 29

---

**User Story 3: Board Management**
As a user,I want an identical Trello home screen, I prefer to manage my projects by creating new boards, seeing all of my current boards, and deleting boards that I no longer need.

Related Github Issues are : 4, 5, 9

---

**User Story 4: List and Card Management**
As a user, I want to manage, monitor, and clean up my tasks in a Kanban-style workflow, I want to add lists to a board, insert task cards into those lists, view the cards under the proper list, and delete lists or cards.

Related Github Issues are : 10, 11, 12, 13, 14, 15, 16, 18, 19, 20, 21, 22, 23

---

**User Story 5: Error Handling and API Responses**
As a user, when something goes wrong, I want the application to provide me with clear error messages and appropriate status codes so I can find out what went poorly and fix it.

Related GitHub Issues are: 30

---

**User Story 6: User Registration and Login**
As a new user, I want to register an account and log in to Taskify so that I can see my boards and tasks are saved to my profile and only I can access them.

Related Github Issues are : 1, 2, 3, 6, 7, 8, 17

---

**User Story 7: Moving Tasks Between Lists**
As a user, I would like to be able to move a task card from one list to another (for example, from To Do to In Progress) so that I could update the status of my work progress, similar to dragging cards in Trello.

Related Github Issues are : No issues created yet, planned for next Sprints.

---

**User Story 8: Real-Time Team Chat**
As a team member, I want to send and receive messages in a real time chat within each project so that my team can discuss tasks and share project updates without leaving the application.

Related Github Issues are : No issues created yet, planned for next Sprints.

---

**User Story 9: Task Details — Descriptions, Labels, and Due Dates**
As a user, In order to make sure that each task has all the information my team needs in one spot, I would like to be able to click on a task card and see a detail page where I can enter a description, select a deadline, and assign color labels.

Related Github Issues: There are no issues created yet; they are planned for the upcoming sprints.

---

**User Story 10: Checklists and Comments on Task Cards**
As a User, In a similar fashion to the card view on Trello, I would like to be able to assign checklists and comments to a task card. I can break down tasks into manageable pieces and communicate with my team regarding the work.

Related Github Issues: There are no issues created yet; they are planned for

---

**User Story 11: Task Assignment for Team Members**
As the project owner,To make it clear who is in responsible for each task,  I would like to attach one or more team members to a task card.

Related Github Issues are  No issues created yet, planned for next Sprints.

---

**User Story 12: Sharing and File Attachments**
To ensure that every necessary document are in the project workspace, I want to exchange files in the project chat and attach documents to task cards as a team member.

Related Github Issues are  No issues created yet, planned for next Sprints.

---

**User Story 13: Search and Filter Tasks**
As a user, To find specific job items across my board fast, I want to be able to search for tasks by keyword and filter them by title, assigned person, or due date.

Related Github Issues are  No issues created yet, planned for next Sprints.

---

**User Story 14: Dashboard and Project Analytics**
As a project owner, To have a high-level picture of the project's advancement without being to go though every card,  I need a dashboard view that displays task counts by stage, past-due tasks, and team workload.

Related Github Issues are  No issues created yet, planned for next Sprints.

---

**User Story 15: Role-Based Access Control**
As a project owner, I want to assign roles (owner, editor, viewer) to team members so that I can control who can create, edit, or only view boards and tasks.

Related Github Issues are  No issues created yet, planned for next Sprints.

---

## Issues to be Resolved in Sprint 1

In Sprint 1, the team's goal was to fix all 30 front-end and back-end problems. Creating a functioning Trello board experience that enables users to manage their workspace from beginning to end and create boards, lists, and task cards was the main goal.

**Backend team (Adithya, Nandhan):** Completely resolved GitHub issues 3, 4, 5, 6, 17, 18, 19, 20, 21, 22, 24, 25, 28, and 30. The implementation of the project structure, database configuration, REST API endpoints for boards, lists, and cards, error management, and user registration and login endpoints were among these challenges.

**Frontend team (Meghana, Sai Sreeja):** GitHub Issues 1, 2, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 23, 26, 27, 29. These included issues related to the implementation of the board page with task cards and list columns, homepage design, routing, responsiveness, and implementation of login and registration pages.

---

## Successfully Completed Issues

We completed **23 out of 30** issues during Sprint 1, covering User Stories 1 through 5. The core Trello-like workflow is fully functional — users can create boards, add lists, add task cards, view everything in a Kanban layout, and delete any item.

### User Story 1: Database, Dependencies, and Project Setup was Finalized

Issues 24, 25, and 28 were finished. A neat folder structure was used to start both the frontend and backend projects. All necessary modules and packages were set up. Project, stage, task, and message tables were included in the database setup, along with the appropriate foreign key restrictions and indexes.

## User Story 2:Homepage, Navigation, and Responsive Design

Issues 5, 26, 27, and 29 were all finished. Every board is shown as a card in a responsive grid on the homepage. In order to facilitate user navigation between the homepage and various board pages, page routing was set up. The arrangement adapts to various screen sizes.

### User Story 3: Board Management

Both issues (4, 9) were completed. Users can create a new board with a name and description, view all existing boards on the homepage, and delete boards they no longer need. A confirmation dialog prevents accidental deletions.

### User Story 4: List and Card Management 
All 13 issues (10, 11, 12, 13, 14, 15, 16, 18, 19, 20, 21,22, 23) were completed. Users can add lists to a board and view them as Kanban columns. Task cards can be created in any list and are shown below the corresponding column. Both lists and cards can be deleted. Issues 13 and 23 were duplicates of 16 and 15, respectively, and were completed as part of the same task.

### User Story 5: Error Handling and API Responses

Issue #30 was completed. The application provides correct status codes and human readable error messages for invalid data, missing resources, and server errors. The frontend shows a fallback view when the backend is unavailable.

### User Story 6: User Registration and Login — Not Completed

None of the seven authentication-related issues were completed during this sprint. Issue 1 (Login page) and 8 (Create Login Component) were not started as the frontend team focused on building the board, list, and card components first. Issue 7 (Create Register Component) was similarly deferred since there was no backend registration logic to connect to. On the backend side, Issue 3 (Implement user registration endpoint) and 17 (Implement POST /login endpoint) were not started because the team prioritized building the project, stage, and task APIs. Issue 6 (Create Authentication Service) was not implemented as there is no auth middleware, JWT handling, or session management yet. Issue 2 (Login page) remained open as the detailed planning for authentication was deferred along with the implementation.

The team decided early on to focus on getting the core board, list, and card workflow working end-to-end before adding user accounts. We felt it was more important to have a fully functional Trello-like board experience first, rather than building login screens without any board functionality behind them. This approach worked well we shipped a complete task management workflow in Sprint 1. All seven authentication issues will carry over to next Sprint as the top priority.

### User Stories 7 to 15: Planned for Future Sprints

User Stories 7 through 15 describe more complex functionality and will be implemented over Sprints 2, 3, and 4. Some of this groundwork has already been done in Sprint 1 the backend has a `PUT /api/tasks/{id}/move` API call for moving tasks between stages (User Story 7), the chat hub and frontend chat service have been implemented using WebSockets (User Story 8), and the data models in the frontend have properties for labels, checklists, comments, attachments, assignees, and due dates (User Stories 9, 10, 11, 12). None of these have been implemented in the UI yet and will require specific issues and development in future sprints.