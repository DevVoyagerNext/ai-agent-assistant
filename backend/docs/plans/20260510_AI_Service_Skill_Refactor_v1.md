# AI Service Skill Handler Refactor Plan (2026-05-10)

## Objective
Refactor the hardcoded `isResumeInterviewSkill` logic in `ai_service.go` into an elegant, maintainable `SkillHandler` architecture. This will encapsulate different AI agents as separate services mapped by `SkillID`.

## Current Pain Points
- `ai_service.go` has a huge `if isResumeInterviewSkill(req.SkillID)` block (lines 662-707).
- Adding new skills/agents will cause `ai_service.go` to grow infinitely with `if-else` or `switch` statements.
- Agent preparation, execution, and fallback logic are coupled with the core `AIChat` session management flow.

## Proposed Architecture

1.  **Define `SkillHandler` Interface (in `service/ai_skill_handler.go`):**
    ```go
    type SkillContext struct {
        Ctx           context.Context
        Req           dto.AIChatReq
        Session       *model.Session
        UserMsg       *model.Message
        AIMsg         *model.Message
        IsNewSession  bool
        PromptContent string
        DB            *gorm.DB
        MsgChan       chan dto.ChatStreamChunk
    }

    type SkillHandler interface {
        CanHandle(skillID string) bool
        Handle(skillCtx *SkillContext, aiService *AIService) // Or aiService can be passed to avoid circular dependencies, or handler has its own deps
    }
    ```

2.  **Implement Default Chat Handler:**
    Move the default `chatAgent` logic (lines 709-813) into a `DefaultChatSkillHandler`.

3.  **Implement Resume Interview Handler:**
    Move the `isResumeInterviewSkill` logic (lines 662-707) into a `ResumeInterviewSkillHandler`.

4.  **Refactor `AIService.AIChat`:**
    - Initialize a list of registered `SkillHandler`s.
    - Iterate through handlers to find one where `CanHandle(req.SkillID) == true`.
    - If found, invoke `Handle()`.
    - If none found, fallback to `DefaultChatSkillHandler`.

## Steps
1. Create `service/ai_skill_handler.go` and define the interface.
2. Create `service/skill_resume_interview.go` and move the resume interview handling logic.
3. Create `service/skill_default_chat.go` and move the standard chat agent logic.
4. Update `ai_service.go` to use the registry pattern to dispatch the request.
5. Clean up `resume_interview_service.go` if necessary to export methods that handlers need.

Please confirm if you agree with this plan before I proceed with the code modifications.
