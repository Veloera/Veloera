# Design Document

## Overview

The inbox system (站内信) is designed as a comprehensive messaging solution that enables administrators to send system notifications to users. The system follows the existing application architecture patterns, utilizing Go with Gin framework for the backend API and React with Semi-UI for the frontend interface.

The system consists of two main components:
1. **User Interface** (`/app/inbox`) - For users to view and manage their received messages
2. **Admin Interface** (`/admin/messages`) - For administrators to create, manage, and send messages

The design emphasizes simplicity, performance, and maintainability while following the established patterns in the codebase.

## Architecture

### Backend Architecture

The backend follows the existing MVC pattern used throughout the application:

```
├── model/
│   ├── message.go          # Message and UserMessage models
│   └── main.go            # Database migration updates
├── controller/
│   ├── message.go         # Message-related API handlers
│   └── user.go           # User message endpoints
├── router/
│   └── api-router.go     # Route definitions
└── dto/
    └── message.go        # Data transfer objects
```

### Frontend Architecture

The frontend follows the existing React structure:

```
web/src/
├── pages/
│   ├── Inbox/            # User inbox interface
│   │   ├── index.js
│   │   └── MessageDetail.js
│   └── Message/          # Admin message management
│       ├── index.js
│       ├── MessageList.js
│       ├── CreateMessage.js
│       └── EditMessage.js
├── components/
│   └── InboxIcon.js      # Header inbox icon component
└── App.js               # Route definitions
```

### Database Schema

The system uses two main tables following the suggested data model:

**messages table:**
```sql
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    format ENUM('markdown', 'html') DEFAULT 'markdown',
    created_at DATETIME NOT NULL,
    updated_at DATETIME,
    created_by INTEGER REFERENCES users(id)
);
```

**user_messages table:**
```sql
CREATE TABLE user_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    message_id INTEGER NOT NULL REFERENCES messages(id),
    read_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    UNIQUE(user_id, message_id)
);
```

## Components and Interfaces

### Backend Components

#### 1. Models (`model/message.go`)

**Message Model:**
```go
type Message struct {
    Id        int       `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title" gorm:"not null"`
    Content   string    `json:"content" gorm:"type:text;not null"`
    Format    string    `json:"format" gorm:"default:markdown"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    CreatedBy int       `json:"created_by"`
}

type UserMessage struct {
    Id        int        `json:"id" gorm:"primaryKey"`
    UserId    int        `json:"user_id" gorm:"not null"`
    MessageId int        `json:"message_id" gorm:"not null"`
    ReadAt    *time.Time `json:"read_at"`
    CreatedAt time.Time  `json:"created_at"`
    Message   Message    `json:"message" gorm:"foreignKey:MessageId"`
}
```

#### 2. Controllers (`controller/message.go`)

**User Message Endpoints:**
- `GetUserMessages(c *gin.Context)` - Get user's messages with pagination
- `MarkMessageAsRead(c *gin.Context)` - Mark specific message as read
- `GetUnreadCount(c *gin.Context)` - Get count of unread messages

**Admin Message Endpoints:**
- `GetAllMessages(c *gin.Context)` - Get all messages with pagination and filters
- `GetMessage(c *gin.Context)` - Get specific message details
- `CreateMessage(c *gin.Context)` - Create and send new message
- `UpdateMessage(c *gin.Context)` - Update existing message
- `DeleteMessage(c *gin.Context)` - Delete message and all user associations

#### 3. DTOs (`dto/message.go`)

```go
type CreateMessageRequest struct {
    Title      string `json:"title" binding:"required"`
    Content    string `json:"content" binding:"required"`
    Format     string `json:"format"`
    UserIds    []int  `json:"user_ids" binding:"required"`
}

type MessageResponse struct {
    Id         int       `json:"id"`
    Title      string    `json:"title"`
    Content    string    `json:"content"`
    Format     string    `json:"format"`
    CreatedAt  time.Time `json:"created_at"`
    IsRead     bool      `json:"is_read"`
    ReadAt     *time.Time `json:"read_at"`
}
```

### Frontend Components

#### 1. User Interface Components

**InboxIcon Component:**
- Displays inbox icon in header next to user avatar
- Shows red badge for unread message count
- Handles click navigation to inbox page

**Inbox Page:**
- Lists all user messages with pagination
- Shows message title, preview, timestamp, and read status
- Supports message detail view with full content rendering
- Handles marking messages as read

#### 2. Admin Interface Components

**Message Management Page:**
- Lists all sent messages with search and filter capabilities
- Provides create, edit, and delete functionality
- Shows recipient count and delivery statistics

**Create/Edit Message Form:**
- Form for composing messages with title and content
- User selection interface (individual or multiple users)
- Format selection (HTML/Markdown)
- Preview functionality

## Data Models

### Message Entity
- **Purpose**: Stores the actual message content and metadata
- **Key Fields**: title, content, format, creation timestamp, creator
- **Relationships**: One-to-many with UserMessage

### UserMessage Entity
- **Purpose**: Associates messages with users and tracks read status
- **Key Fields**: user_id, message_id, read_at timestamp
- **Relationships**: Many-to-one with User and Message

### Data Flow
1. Admin creates message and selects recipients
2. System creates Message record
3. System creates UserMessage records for each recipient
4. Users see unread count in header badge
5. Users view messages in inbox, marking them as read
6. System updates read_at timestamp in UserMessage

## Error Handling

### Backend Error Handling
- **Validation Errors**: Return 400 with descriptive error messages
- **Authorization Errors**: Return 403 for insufficient permissions
- **Not Found Errors**: Return 404 for non-existent resources
- **Database Errors**: Return 500 with generic error message, log details
- **Concurrent Access**: Handle race conditions in read status updates

### Frontend Error Handling
- **API Errors**: Display user-friendly error messages using Semi-UI Toast
- **Network Errors**: Show retry options and offline indicators
- **Loading States**: Provide loading spinners and skeleton screens
- **Empty States**: Show appropriate messages when no data exists

### Error Recovery
- **Graceful Degradation**: System continues to function if messaging is unavailable
- **Retry Logic**: Automatic retry for transient failures
- **Fallback UI**: Basic functionality available even with partial failures

## Testing Strategy

### Backend Testing
- **Unit Tests**: Test individual model methods and controller functions
- **Integration Tests**: Test API endpoints with database interactions
- **Performance Tests**: Verify pagination and query performance with large datasets

### Frontend Testing
- **Component Tests**: Test individual React components in isolation
- **Integration Tests**: Test user workflows and API interactions
- **E2E Tests**: Test complete user journeys from login to message interaction

### Test Data
- **Mock Data**: Generate realistic test messages and user associations
- **Edge Cases**: Test with empty states, large content, and special characters
- **Performance Data**: Test with thousands of messages and users

## Security Considerations

### Authentication & Authorization
- **User Authentication**: Verify user identity for all message operations
- **Admin Authorization**: Restrict message management to admin users only
- **Message Access**: Users can only access their own messages

### Data Protection
- **Input Validation**: Sanitize all user inputs to prevent XSS and injection
- **Content Security**: Validate HTML content and restrict dangerous elements
- **Rate Limiting**: Prevent spam and abuse through rate limiting

### Privacy
- **Message Privacy**: Messages are only visible to intended recipients
- **Audit Trail**: Log message creation and access for security monitoring
- **Data Retention**: Consider implementing message retention policies

## Performance Optimization

### Database Optimization
- **Indexing**: Create indexes on user_id, message_id, and read_at columns
- **Pagination**: Implement efficient pagination for large message lists
- **Query Optimization**: Use joins to minimize database round trips

### Caching Strategy
- **Unread Counts**: Cache unread message counts per user
- **Message Content**: Cache frequently accessed messages
- **User Lists**: Cache user lists for admin message creation

### Frontend Optimization
- **Lazy Loading**: Load message content on demand
- **Virtual Scrolling**: Handle large message lists efficiently
- **Code Splitting**: Split inbox and admin interfaces into separate bundles

## Scalability Considerations

### Database Scaling
- **Partitioning**: Consider partitioning user_messages by user_id for large datasets
- **Archiving**: Implement message archiving for old messages
- **Read Replicas**: Use read replicas for message retrieval operations

### Application Scaling
- **Stateless Design**: Ensure all components are stateless for horizontal scaling
- **Background Processing**: Use queues for bulk message sending operations
- **CDN Integration**: Serve static assets through CDN

### Monitoring & Metrics
- **Performance Metrics**: Track message delivery times and read rates
- **Usage Analytics**: Monitor message engagement and user behavior
- **Error Tracking**: Implement comprehensive error logging and alerting