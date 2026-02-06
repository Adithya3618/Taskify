# Taskify (Angular Frontend)

A kanban board application built with **Angular 17** and **Angular CDK** for drag-and-drop. Frontend only; data is kept in memory.

## Features

- **Board** – Single board with a header and horizontal list layout
- **Lists** – Add/delete lists (columns) such as To Do, In Progress, Done
- **Cards** – Add/edit/delete cards; inline edit title
- **Drag & drop** – Move cards within a list (reorder) or between lists using Angular CDK
- **Taskify UI** – Blue header, gray list columns, white cards

## Tech stack

- Angular 17 (standalone components, signals, computed)
- Angular CDK (DragDropModule)
- SCSS
- In-memory state (no backend)

## Setup

1. **Install dependencies** (if not already done):

   ```bash
   cd taskify
   npm install
   ```

2. **Run the dev server**:

   ```bash
   npm start
   ```

3. Open **http://localhost:4200/** in your browser.

## Build

```bash
npm run build
```

Output is in `dist/taskify/browser/`.

## Project structure

- `src/app/models/` – Board, List, Card interfaces
- `src/app/services/board.service.ts` – In-memory state and CRUD for boards, lists, cards
- `src/app/components/board/` – Board view and “Add list”
- `src/app/components/list/` – List column and “Add card”
- `src/app/components/card/` – Card with edit/delete

## Note

If you see `Cannot find module 'rxjs'` when running `ng build` or `ng serve`, run a clean install:

```bash
Remove-Item -Recurse -Force node_modules
Remove-Item -Force package-lock.json
npm install
```

Then run `npm start` or `npm run build` again.
