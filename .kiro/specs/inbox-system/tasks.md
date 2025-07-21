# Implementation Plan

- [x] 1. Set up database models and migrations

  - Create Message and UserMessage models in Go with proper GORM tags
  - Add database migration logic to model/main.go for the new tables
  - Implement model methods for CRUD operations and message queries
  - _Requirements: 7.1, 7.2, 7.3_

- [x] 2. Implement backend API endpoints for user message functionality

  - Create controller/message.go with user-facing message endpoints
  - Implement GetUserMessages endpoint with pagination and read status
  - Implement MarkMessageAsRead endpoint for updating read timestamps
  - Implement GetUnreadCount endpoint for header badge display
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 8.1, 8.2_

- [x] 3. Implement backend API endpoints for admin message management

  - Add admin message management endpoints to controller/message.go
  - Implement GetAllMessages endpoint with search and filter capabilities
  - Implement CreateMessage endpoint for sending messages to multiple users
  - Implement UpdateMessage and DeleteMessage endpoints with proper authorization
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 8.3, 8.4, 8.5, 8.6, 8.7, 8.8_

- [x] 4. Add API routes and middleware configuration

  - Update router/api-router.go to include message-related routes
  - Configure proper authentication middleware for user and admin endpoints
  - Set up route grouping for /api/user/messages and /api/admin/messages
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_

- [x] 5. Create user inbox interface components

  - Create web/src/pages/Inbox/index.js for the main inbox page
  - Implement message list display with title, preview, timestamp, and read status
  - Create MessageDetail component for full message content viewing
  - Add support for HTML and Markdown content rendering
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_

- [x] 6. Implement inbox header icon and notification badge

  - Create web/src/components/InboxIcon.js component for header display
  - Add inbox icon next to user avatar in the header layout
  - Implement red badge indicator for unread message count
  - Add click navigation to /app/inbox route
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 7. Create admin message management interface

  - Create web/src/pages/Message/index.js for admin message list
  - Implement MessageList component with search, filter, and pagination
  - Create CreateMessage component for composing and sending new messages
  - Create EditMessage component for updating existing messages
  - Add user selection interface for choosing message recipients
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 8. Add frontend routing and navigation

  - Update web/src/App.js to include /app/inbox and /admin/messages routes
  - Configure proper route protection for user and admin interfaces
  - Add navigation menu items for admin message management
  - _Requirements: 1.1, 4.1, 5.1_

- [ ] 9. Implement real-time unread count updates

  - Add API polling or WebSocket support for real-time badge updates
  - Update InboxIcon component to refresh unread count periodically
  - Handle read status changes to update badge count immediately
  - _Requirements: 2.4, 3.4, 3.5_

- [-] 10. Add comprehensive error handling and validation

  - Implement input validation for all API endpoints
  - Add proper error responses with descriptive messages
  - Create frontend error handling with user-friendly notifications
  - Add loading states and empty state handling for all components
  - _Requirements: 5.6, 6.6, 8.8_

- [ ] 11. Create unit tests for backend functionality

  - Write unit tests for Message and UserMessage model methods
  - Create tests for all controller endpoints with various scenarios
  - Test authentication and authorization for admin endpoints
  - Add tests for edge cases like concurrent read operations
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7, 8.8_

- [ ] 12. Integrate and test complete user workflow
  - Test end-to-end user flow from message creation to reading
  - Verify admin can create messages and assign to multiple users
  - Test user inbox displays messages correctly with proper read status
  - Validate header badge updates when messages are read/unread
  - Test responsive design and cross-browser compatibility
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_
