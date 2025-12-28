"""
Example worker tasks using generated types.

This demonstrates how the same OpenAPI types are shared between:
- Frontend (TypeScript)
- API service (Go)
- Worker service (Python)
"""

from .types.api import Document, DocumentStatus, User


def process_document_publish(document: Document) -> None:
    """
    Process a document when it's published.

    The document parameter is typed using the generated TypedDict,
    ensuring type safety across all services.
    """
    if document.get("status") != DocumentStatus.PUBLISHED:
        return

    author_id = document["authorId"]
    title = document["title"]

    # Example: send notification, index for search, etc.
    print(f"Document '{title}' published by author {author_id}")


def notify_user(user: User, message: str) -> None:
    """
    Send a notification to a user.

    Uses the shared User type from the generated types.
    """
    email = user["email"]
    name = user["name"]

    print(f"Notifying {name} ({email}): {message}")


def check_workspace_limits(workspace_id: str, member_count: int) -> bool:
    """
    Check if a workspace has reached its member limit.

    This would use the PlanType enum to determine limits.
    """
    # Example business logic using shared types
    return member_count < 100

