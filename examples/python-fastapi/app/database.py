"""
In-memory database simulation for the example.

In a real application, you would use SQLAlchemy, Tortoise ORM, or similar.
This module demonstrates how generated types flow through your data layer.
"""

from datetime import datetime, date
from uuid import UUID, uuid4
from typing import TypedDict

from .types import Task, Project, User, TaskStatus, Priority


# =============================================================================
# Type definitions for database records
# =============================================================================


class TaskRecord(TypedDict):
    """Internal database record for a task."""

    id: str
    title: str
    description: str | None
    status: TaskStatus
    priority: Priority
    assignee_id: str | None
    project_id: str | None
    labels: list[str]
    due_date: str | None
    estimated_hours: float | None
    completed_at: str | None
    created_at: str
    updated_at: str | None


class ProjectRecord(TypedDict):
    """Internal database record for a project."""

    id: str
    name: str
    description: str | None
    color: str | None
    owner_id: str | None
    created_at: str


class UserRecord(TypedDict):
    """Internal database record for a user."""

    id: str
    email: str
    name: str
    avatar_url: str | None


# =============================================================================
# In-memory storage
# =============================================================================

_tasks: dict[str, TaskRecord] = {}
_projects: dict[str, ProjectRecord] = {}
_users: dict[str, UserRecord] = {}


def _init_sample_data() -> None:
    """Initialize with sample data."""
    # Sample users
    user1_id = str(uuid4())
    user2_id = str(uuid4())

    _users[user1_id] = {
        "id": user1_id,
        "email": "alice@example.com",
        "name": "Alice Johnson",
        "avatar_url": "https://api.dicebear.com/7.x/avataaars/svg?seed=alice",
    }
    _users[user2_id] = {
        "id": user2_id,
        "email": "bob@example.com",
        "name": "Bob Smith",
        "avatar_url": "https://api.dicebear.com/7.x/avataaars/svg?seed=bob",
    }

    # Sample project
    project_id = str(uuid4())
    _projects[project_id] = {
        "id": project_id,
        "name": "Website Redesign",
        "description": "Redesign the company website",
        "color": "#3B82F6",
        "owner_id": user1_id,
        "created_at": datetime.now().isoformat(),
    }

    # Sample tasks
    task1_id = str(uuid4())
    _tasks[task1_id] = {
        "id": task1_id,
        "title": "Design new homepage mockup",
        "description": "Create wireframes and high-fidelity mockups for the new homepage",
        "status": TaskStatus.IN_PROGRESS,
        "priority": Priority.HIGH,
        "assignee_id": user1_id,
        "project_id": project_id,
        "labels": ["design", "frontend"],
        "due_date": "2024-02-15",
        "estimated_hours": 16.0,
        "completed_at": None,
        "created_at": datetime.now().isoformat(),
        "updated_at": None,
    }

    task2_id = str(uuid4())
    _tasks[task2_id] = {
        "id": task2_id,
        "title": "Set up CI/CD pipeline",
        "description": None,
        "status": TaskStatus.TODO,
        "priority": Priority.MEDIUM,
        "assignee_id": user2_id,
        "project_id": project_id,
        "labels": ["devops"],
        "due_date": None,
        "estimated_hours": 8.0,
        "completed_at": None,
        "created_at": datetime.now().isoformat(),
        "updated_at": None,
    }


# Initialize sample data on module load
_init_sample_data()


# =============================================================================
# Database functions returning generated types
# =============================================================================


def _record_to_user(record: UserRecord) -> User:
    """Convert a database record to the generated User type."""
    return User(
        id=record["id"],
        email=record["email"],
        name=record["name"],
        avatar_url=record.get("avatar_url"),
    )


def _record_to_task(record: TaskRecord) -> Task:
    """Convert a database record to the generated Task type."""
    assignee = None
    if record["assignee_id"] and record["assignee_id"] in _users:
        assignee = _record_to_user(_users[record["assignee_id"]])

    return Task(
        id=record["id"],
        title=record["title"],
        description=record.get("description"),
        status=record["status"],
        priority=record["priority"],
        assignee=assignee,
        project_id=record.get("project_id"),
        labels=record.get("labels", []),
        due_date=record.get("due_date"),
        estimated_hours=record.get("estimated_hours"),
        completed_at=record.get("completed_at"),
        created_at=record["created_at"],
        updated_at=record.get("updated_at"),
    )


def _record_to_project(record: ProjectRecord) -> Project:
    """Convert a database record to the generated Project type."""
    owner = None
    if record["owner_id"] and record["owner_id"] in _users:
        owner = _record_to_user(_users[record["owner_id"]])

    task_count = sum(1 for t in _tasks.values() if t["project_id"] == record["id"])

    return Project(
        id=record["id"],
        name=record["name"],
        description=record.get("description"),
        color=record.get("color"),
        task_count=task_count,
        owner=owner,
        created_at=record["created_at"],
    )


# =============================================================================
# CRUD operations
# =============================================================================


def get_all_tasks(
    status: TaskStatus | None = None,
    priority: Priority | None = None,
    assignee_id: UUID | None = None,
) -> list[Task]:
    """Get all tasks, optionally filtered."""
    results: list[Task] = []

    for record in _tasks.values():
        if status and record["status"] != status:
            continue
        if priority and record["priority"] != priority:
            continue
        if assignee_id and record["assignee_id"] != str(assignee_id):
            continue
        results.append(_record_to_task(record))

    return results


def get_task(task_id: UUID) -> Task | None:
    """Get a task by ID."""
    record = _tasks.get(str(task_id))
    if record:
        return _record_to_task(record)
    return None


def create_task(
    title: str,
    description: str | None = None,
    priority: Priority = Priority.MEDIUM,
    assignee_id: UUID | None = None,
    project_id: UUID | None = None,
    labels: list[str] | None = None,
    due_date: date | None = None,
    estimated_hours: float | None = None,
) -> Task:
    """Create a new task."""
    task_id = str(uuid4())
    now = datetime.now().isoformat()

    record: TaskRecord = {
        "id": task_id,
        "title": title,
        "description": description,
        "status": TaskStatus.TODO,
        "priority": priority,
        "assignee_id": str(assignee_id) if assignee_id else None,
        "project_id": str(project_id) if project_id else None,
        "labels": labels or [],
        "due_date": due_date.isoformat() if due_date else None,
        "estimated_hours": estimated_hours,
        "completed_at": None,
        "created_at": now,
        "updated_at": None,
    }

    _tasks[task_id] = record
    return _record_to_task(record)


def update_task(task_id: UUID, **updates: object) -> Task | None:
    """Update a task."""
    record = _tasks.get(str(task_id))
    if not record:
        return None

    for key, value in updates.items():
        if value is not None and key in record:
            if key == "due_date" and isinstance(value, date):
                record[key] = value.isoformat()  # type: ignore
            elif key == "assignee_id" and isinstance(value, UUID):
                record[key] = str(value)  # type: ignore
            else:
                record[key] = value  # type: ignore

    record["updated_at"] = datetime.now().isoformat()

    # Auto-set completed_at when status changes to done
    if updates.get("status") == TaskStatus.DONE and not record["completed_at"]:
        record["completed_at"] = datetime.now().isoformat()

    return _record_to_task(record)


def delete_task(task_id: UUID) -> bool:
    """Delete a task."""
    if str(task_id) in _tasks:
        del _tasks[str(task_id)]
        return True
    return False


def get_all_projects() -> list[Project]:
    """Get all projects."""
    return [_record_to_project(r) for r in _projects.values()]


def create_project(
    name: str,
    description: str | None = None,
    color: str | None = None,
) -> Project:
    """Create a new project."""
    project_id = str(uuid4())
    now = datetime.now().isoformat()

    record: ProjectRecord = {
        "id": project_id,
        "name": name,
        "description": description,
        "color": color,
        "owner_id": None,  # Would be set from authenticated user
        "created_at": now,
    }

    _projects[project_id] = record
    return _record_to_project(record)


def get_all_users() -> list[User]:
    """Get all users."""
    return [_record_to_user(r) for r in _users.values()]

