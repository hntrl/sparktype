"""
Pydantic models that extend the generated TypedDict types.

This demonstrates how to use sparktype's generated TypedDicts as a foundation
for Pydantic models, which add validation and serialization capabilities.

Pattern:
1. sparktype generates TypedDict classes (static type hints)
2. We create Pydantic models that match the same structure
3. TypedDict is used for function signatures and type hints
4. Pydantic models are used for request validation and response serialization
"""

from datetime import date, datetime
from uuid import UUID

from pydantic import BaseModel, Field, EmailStr, HttpUrl

from .types import Priority, TaskStatus


# =============================================================================
# User Models
# =============================================================================


class UserModel(BaseModel):
    """Pydantic model matching the generated User TypedDict."""

    id: UUID
    email: EmailStr
    name: str
    avatar_url: HttpUrl | None = None


# =============================================================================
# Task Models
# =============================================================================


class CreateTaskModel(BaseModel):
    """
    Pydantic model for task creation requests.

    This validates incoming request data and can be converted to the
    CreateTaskRequest TypedDict for use with typed functions.
    """

    title: str = Field(min_length=1, max_length=200)
    description: str | None = None
    priority: Priority = Priority.MEDIUM
    assignee_id: UUID | None = None
    project_id: UUID | None = None
    labels: list[str] = Field(default_factory=list)
    due_date: date | None = None
    estimated_hours: float | None = Field(default=None, ge=0)


class UpdateTaskModel(BaseModel):
    """Pydantic model for task update requests."""

    title: str | None = Field(default=None, min_length=1, max_length=200)
    description: str | None = None
    status: TaskStatus | None = None
    priority: Priority | None = None
    assignee_id: UUID | None = None
    labels: list[str] | None = None
    due_date: date | None = None
    estimated_hours: float | None = Field(default=None, ge=0)


class TaskModel(BaseModel):
    """Pydantic model matching the generated Task TypedDict."""

    id: UUID
    title: str
    description: str | None = None
    status: TaskStatus
    priority: Priority
    assignee: UserModel | None = None
    project_id: UUID | None = None
    labels: list[str] = Field(default_factory=list)
    due_date: date | None = None
    estimated_hours: float | None = None
    completed_at: datetime | None = None
    created_at: datetime
    updated_at: datetime | None = None


# =============================================================================
# Project Models
# =============================================================================


class CreateProjectModel(BaseModel):
    """Pydantic model for project creation requests."""

    name: str = Field(min_length=1, max_length=100)
    description: str | None = None
    color: str | None = Field(default=None, pattern=r"^#[0-9A-Fa-f]{6}$")


class ProjectModel(BaseModel):
    """Pydantic model matching the generated Project TypedDict."""

    id: UUID
    name: str
    description: str | None = None
    color: str | None = None
    task_count: int = 0
    owner: UserModel | None = None
    created_at: datetime


# =============================================================================
# Error Models
# =============================================================================


class ErrorResponseModel(BaseModel):
    """Pydantic model matching the generated ErrorResponse TypedDict."""

    error: str
    message: str
    request_id: str | None = None

