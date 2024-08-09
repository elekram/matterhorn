# Matterhorn.

Web app pattern using vanilla Golang for the back end. This project is ready to pull a then extend. Bootstrap and HTMX 2 are integrated and ready to use. This project needs and uses Docker Compose.

### Overview

**Includes a dependency Injected server struct with the following:**
- Session Management Middleware (cookies and inbuilt memory store with option to add addtional stores using a DB and store interface)
- Middleware to set headers for cache and security
- Sign-in page utilising OAuth that redirects to an app base page ready to be extended
- HTMX 2
- Bootstrap 5
- Mongo DB
- Logging
- Configuration management for os environment variables

Named after the mountain