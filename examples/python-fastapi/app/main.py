"""
FastAPI application demonstrating sparktype generated types.

This shows how to build a type-safe API where:
1. Request bodies are validated using Pydantic models
2. Response bodies match the generated TypedDict types
3. Type hints flow through the entire application
"""

from uuid import UUID

from fastapi import FastAPI, HTTPException, status
from fastapi.responses import JSONResponse

from .types import Task, Project, User, TaskStatus, Priority, ErrorResponse
from .models import (
    CreateTaskModel,
    UpdateTaskModel,
    TaskModel,
    CreateProjectModel,
    ProjectModel,
    UserModel,
    ErrorResponseModel,
)
from . import database as db


# =============================================================================
# Application setup
# =============================================================================

app = FastAPI(
    title="Task Management API",
    description="Example API demonstrating sparktype with FastAPI",
    version="1.0.0",
)


# =============================================================================
# Exception handlers
# =============================================================================


@app.exception_handler(HTTPException)
async def http_exception_handler(request, exc: HTTPException):  # type: ignore
    """Return errors in the ErrorResponse format from our OpenAPI spec."""
    error_response: ErrorResponse = {
        "error": exc.detail if isinstance(exc.detail, str) else "error",
        "message": str(exc.detail),
    }
    return JSONResponse(status_code=exc.status_code, content=error_response)


# =============================================================================
# Task endpoints
# =============================================================================


@app.get("/tasks", response_model=list[TaskModel])
async def list_tasks(
    status: TaskStatus | None = None,
    priority: Priority | None = None,
    assignee_id: UUID | None = None,
) -> list[Task]:
    """
    List all tasks with optional filters.

    The return type `list[Task]` uses the generated TypedDict,
    while `response_model=list[TaskModel]` handles serialization.
    """
    return db.get_all_tasks(
        status=status,
        priority=priority,
        assignee_id=assignee_id,
    )


@app.post("/tasks", response_model=TaskModel, status_code=status.HTTP_201_CREATED)
async def create_task(request: CreateTaskModel) -> Task:
    """
    Create a new task.

    The request body is validated by Pydantic (CreateTaskModel).
    The function returns a Task TypedDict for type safety.
    """
    return db.create_task(
        title=request.title,
        description=request.description,
        priority=request.priority,
        assignee_id=request.assignee_id,
        project_id=request.project_id,
        labels=request.labels,
        due_date=request.due_date,
        estimated_hours=request.estimated_hours,
    )


@app.get("/tasks/{task_id}", response_model=TaskModel)
async def get_task(task_id: UUID) -> Task:
    """Get a task by ID."""
    task = db.get_task(task_id)
    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Task not found",
        )
    return task


@app.patch("/tasks/{task_id}", response_model=TaskModel)
async def update_task(task_id: UUID, request: UpdateTaskModel) -> Task:
    """
    Update a task.

    Only fields that are provided (not None) will be updated.
    This pattern uses Pydantic's exclude_unset to handle partial updates.
    """
    # Get only the fields that were explicitly set in the request
    updates = request.model_dump(exclude_unset=True)

    task = db.update_task(task_id, **updates)
    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Task not found",
        )
    return task


@app.delete("/tasks/{task_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_task(task_id: UUID) -> None:
    """Delete a task."""
    if not db.delete_task(task_id):
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Task not found",
        )


# =============================================================================
# Project endpoints
# =============================================================================


@app.get("/projects", response_model=list[ProjectModel])
async def list_projects() -> list[Project]:
    """List all projects."""
    return db.get_all_projects()


@app.post(
    "/projects", response_model=ProjectModel, status_code=status.HTTP_201_CREATED
)
async def create_project(request: CreateProjectModel) -> Project:
    """Create a new project."""
    return db.create_project(
        name=request.name,
        description=request.description,
        color=request.color,
    )


# =============================================================================
# User endpoints
# =============================================================================


@app.get("/users", response_model=list[UserModel])
async def list_users() -> list[User]:
    """List all users."""
    return db.get_all_users()


# =============================================================================
# Health check
# =============================================================================


@app.get("/health")
async def health_check() -> dict[str, str]:
    """Health check endpoint."""
    return {"status": "healthy"}

