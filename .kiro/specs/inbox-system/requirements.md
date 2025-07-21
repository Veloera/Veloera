# Requirements Document

## Introduction

This feature implements a comprehensive in-app messaging system (站内信) for the application, providing system notifications to users. The system consists of two main components: a user-facing inbox interface for viewing and managing received messages, and an administrative interface for creating, managing, and sending messages to users. The system is designed for one-way communication from administrators to users, focusing on system notifications rather than user-to-user messaging.

## Requirements

### Requirement 1

**User Story:** As a user, I want to view all my received system messages in an organized inbox, so that I can stay informed about important system notifications and updates.

#### Acceptance Criteria

1. WHEN a user navigates to `/app/inbox` THEN the system SHALL display a list of all messages sent to that user
2. WHEN displaying messages THEN the system SHALL show the title, content preview, timestamp, and read status for each message
3. WHEN a message is unread THEN the system SHALL visually distinguish it from read messages
4. WHEN a user clicks on a message THEN the system SHALL display the full message content
5. WHEN message content is in HTML format THEN the system SHALL render it as HTML
6. WHEN message content is in Markdown format THEN the system SHALL render it as formatted text

### Requirement 2

**User Story:** As a user, I want to mark messages as read and see visual indicators for unread messages, so that I can easily track which notifications I have already reviewed.

#### Acceptance Criteria

1. WHEN a user opens a message THEN the system SHALL automatically mark it as read
2. WHEN a user has unread messages THEN the system SHALL display a notification badge on the inbox icon
3. WHEN a message is marked as read THEN the system SHALL update the read timestamp
4. WHEN displaying the inbox THEN the system SHALL show the total count of unread messages
5. IF a user manually marks a message as read THEN the system SHALL update the read status immediately

### Requirement 3

**User Story:** As a user, I want to easily access my inbox from anywhere in the application, so that I can quickly check for new notifications without navigating through multiple pages.

#### Acceptance Criteria

1. WHEN a user is on any page THEN the system SHALL display an inbox icon in the header next to the user avatar
2. WHEN a user has unread messages THEN the system SHALL display a red badge indicator on the inbox icon
3. WHEN a user clicks the inbox icon THEN the system SHALL navigate to `/app/inbox`
4. WHEN the badge count changes THEN the system SHALL update the display in real-time
5. WHEN a user has no unread messages THEN the system SHALL hide the badge indicator

### Requirement 4

**User Story:** As an administrator, I want to view and manage all system messages that have been sent, so that I can track communication history and maintain message records.

#### Acceptance Criteria

1. WHEN an administrator navigates to `/admin/messages` THEN the system SHALL display a list of all sent messages
2. WHEN displaying the message list THEN the system SHALL show title, recipient count, creation date, and delivery status
3. WHEN an administrator wants to filter messages THEN the system SHALL provide search and filter options by title, recipient, and date
4. WHEN an administrator clicks on a message THEN the system SHALL display detailed message information
5. WHEN viewing message details THEN the system SHALL show recipient list and read status for each recipient

### Requirement 5

**User Story:** As an administrator, I want to create and send new system messages to specific users or groups, so that I can communicate important information and updates effectively.

#### Acceptance Criteria

1. WHEN an administrator navigates to `/admin/messages/new` THEN the system SHALL display a message creation form
2. WHEN creating a message THEN the system SHALL require a title and content
3. WHEN selecting recipients THEN the system SHALL allow choosing individual users or multiple users
4. WHEN composing content THEN the system SHALL support both HTML and Markdown formats
5. WHEN ready to send THEN the system SHALL provide an option to send immediately
6. WHEN a message is sent THEN the system SHALL create message records for all specified recipients
7. IF message creation fails THEN the system SHALL display appropriate error messages

### Requirement 6

**User Story:** As an administrator, I want to edit and delete existing messages, so that I can correct mistakes and manage outdated information.

#### Acceptance Criteria

1. WHEN an administrator views a message THEN the system SHALL provide edit and delete options
2. WHEN editing a message THEN the system SHALL allow modification of title, content, and format
3. WHEN a message is edited THEN the system SHALL update the message for all recipients
4. WHEN deleting a message THEN the system SHALL remove it from all recipients' inboxes
5. WHEN performing edit or delete operations THEN the system SHALL require confirmation
6. IF edit or delete operations fail THEN the system SHALL display appropriate error messages

### Requirement 7

**User Story:** As a system, I want to maintain proper data relationships and integrity for messages and user associations, so that the messaging system operates reliably and efficiently.

#### Acceptance Criteria

1. WHEN a message is created THEN the system SHALL store it in the messages table with unique ID
2. WHEN assigning messages to users THEN the system SHALL create user_message associations
3. WHEN a user reads a message THEN the system SHALL update the read_at timestamp in user_message table
4. WHEN querying user messages THEN the system SHALL efficiently join message and user_message tables
5. WHEN a user is deleted THEN the system SHALL handle associated message records appropriately
6. WHEN a message is deleted THEN the system SHALL remove all associated user_message records

### Requirement 8

**User Story:** As a developer, I want the system to provide proper API endpoints for both user and admin functionality, so that the frontend can interact with the messaging system effectively.

#### Acceptance Criteria

1. WHEN frontend requests user messages THEN the system SHALL provide GET `/api/user/messages` endpoint
2. WHEN marking messages as read THEN the system SHALL provide PUT `/api/user/messages/:id/read` endpoint
3. WHEN admin requests message list THEN the system SHALL provide GET `/api/admin/messages` endpoint
4. WHEN creating new messages THEN the system SHALL provide POST `/api/admin/messages` endpoint
5. WHEN updating messages THEN the system SHALL provide PUT `/api/admin/messages/:id` endpoint
6. WHEN deleting messages THEN the system SHALL provide DELETE `/api/admin/messages/:id` endpoint
7. WHEN accessing admin endpoints THEN the system SHALL verify administrator privileges
8. WHEN API operations fail THEN the system SHALL return appropriate HTTP status codes and error messages