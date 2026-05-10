# AI Module Refactoring Plan (2026-05-10)

## Objective
Refactor the AI-related codebase by organizing files into dedicated package directories for better maintainability and module cohesion, following the user's instructions to create `aiservice` and `aidao` folders.

## Current State
All AI-related services are currently mixed in the `service/` directory, and there is no dedicated DAO for AI database operations (currently mixed inside the service layer).

## Action Plan

### 1. Create New Directories
- Create `service/aiservice/`
- Create `dao/aidao/`

### 2. Move & Rename Service Files
Move the following AI-related files from `service/` to `service/aiservice/` and change their package name from `package service` to `package aiservice`:
- `ai_service.go`
- `ai_skill_handler.go`
- `ai_tool_service.go`
- `resume_interview_service.go`
- `skill_default_chat.go`
- `skill_resume_interview.go`

### 3. Create AI DAO
- Create a new file `ai_dao.go` inside `dao/aidao/` with `package aidao`.
- Extract common AI database operations (like creating sessions, saving messages, updating AI replies) from `ai_service.go` into `AIDao` to properly decouple the database layer.

### 4. Update Dependencies & Imports
- Update `controller/ai_controller.go` to import and use the new `aiservice.AIService`.
- Update the `router` package (likely `router/ai_router.go` or `router/router.go`) if it directly instantiates the AI service or controller.
- Update `AIService` in `service/aiservice/` to use the newly created `aidao.AIDao` for database interactions instead of directly using `global.GVA_DB` or passing `gorm.DB` around directly where possible, or just initialize it.

## Next Steps
Once you approve this plan, I will execute these changes, update the imports, and ensure the project compiles successfully.
