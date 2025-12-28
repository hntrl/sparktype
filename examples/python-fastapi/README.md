# Python + FastAPI Example

This example demonstrates how to use sparktype in a FastAPI backend for type-safe API development.

## What This Example Shows

- **Generated TypedDicts** - Type hints for request/response bodies
- **Pydantic integration** - Using generated types alongside Pydantic models
- **Enum types** - Type-safe status and priority values
- **Database layer** - How types flow through your data layer
- **mypy compatibility** - Static type checking with generated types

## Project Structure

```
python-fastapi/
├── openapi.yaml          # Task Management API spec
├── typegen.jsonc         # sparktype configuration
├── pyproject.toml        # Python project configuration
├── app/
│   ├── __init__.py
│   ├── main.py           # FastAPI application
│   ├── models.py         # Pydantic models
│   ├── database.py       # Data layer with typed functions
│   └── types/
│       ├── __init__.py
│       └── api.py        # Generated TypedDict classes
```

## Getting Started

### 1. Create virtual environment

```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

### 2. Install dependencies

```bash
pip install -e ".[dev]"
```

### 3. Generate types

```bash
sparktype generate
```

This generates `app/types/api.py` with TypedDict classes for all schemas.

### 4. Run the server

```bash
uvicorn app.main:app --reload
```

Visit http://localhost:8000/docs for the interactive API documentation.

## How It Works

### Generated TypedDicts

sparktype generates Python TypedDicts that provide static type hints:

```python
# Generated in app/types/api.py
class Task(TypedDict):
    """A task in the system"""
    id: str
    title: str
    description: NotRequired[str]
    status: TaskStatus
    priority: Priority
    # ...

class TaskStatus(str, Enum):
    """Task status"""
    TODO = "todo"
    IN_PROGRESS = "in_progress"
    REVIEW = "review"
    DONE = "done"
    CANCELLED = "cancelled"
```

### Using Types in FastAPI

The generated types work seamlessly with FastAPI:

```python
from .types import Task, TaskStatus, Priority

@app.get("/tasks", response_model=list[TaskModel])
async def list_tasks(
    status: TaskStatus | None = None,  # Enum type from spec
    priority: Priority | None = None,
) -> list[Task]:  # Return type uses generated TypedDict
    return db.get_all_tasks(status=status, priority=priority)
```

### Type-Safe Database Layer

Types flow through your entire application:

```python
def get_task(task_id: UUID) -> Task | None:
    """Get a task by ID - returns the generated Task type."""
    record = _tasks.get(str(task_id))
    if record:
        return Task(
            id=record["id"],
            title=record["title"],
            status=record["status"],  # TaskStatus enum
            # ... all fields are type-checked
        )
    return None
```

## Type Checking

Run mypy to verify types:

```bash
mypy app
```

The generated types are compatible with mypy's strict mode.

## CI Integration

Add type checking and type generation to your CI pipeline:

```yaml
# .github/workflows/ci.yml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      
      - name: Install dependencies
        run: pip install -e ".[dev]"
      
      - name: Check generated types
        run: sparktype check
      
      - name: Type check
        run: mypy app
      
      - name: Run tests
        run: pytest
```

## Key Patterns

### 1. TypedDict for Type Safety

Use generated TypedDicts for function signatures:

```python
def process_task(task: Task) -> None:
    # task.status is typed as TaskStatus
    if task["status"] == TaskStatus.DONE:
        send_notification(task["assignee"])
```

### 2. Enums for Constrained Values

The generated enums provide type-safe constants:

```python
from .types import TaskStatus, Priority

# Type checker ensures valid values
task["status"] = TaskStatus.IN_PROGRESS  # ✓
task["status"] = "invalid"  # ✗ Type error
```

### 3. Optional Fields

NotRequired fields are properly typed:

```python
# description is NotRequired[str]
if task.get("description"):
    print(task["description"])  # Type narrowed to str
```

## Testing

The example includes type-safe test patterns:

```python
import pytest
from httpx import AsyncClient
from app.main import app
from app.types import Task, TaskStatus

@pytest.mark.asyncio
async def test_create_task():
    async with AsyncClient(app=app, base_url="http://test") as client:
        response = await client.post("/tasks", json={
            "title": "Test task",
            "priority": "high",
        })
        assert response.status_code == 201
        
        # Response matches Task type
        task: Task = response.json()
        assert task["status"] == TaskStatus.TODO
```

