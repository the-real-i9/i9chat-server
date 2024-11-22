# i9chat (API Server)

## Overview

i9chat is a Websocket-based API server for a chat application, designed as a portfolio project to showcase my backend development skills. It is aimed at fellow backend engineers and hiring managers seeking highly skilled developers.

>⚠️ **Important Note:**\
>This API server is built entirely with WebSockets,meaning all endpoints use WebSocket communication and require an active WebSocket connection.

## Table of Contents

- [Overview](#overview)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Diagrams](#diagrams): ER, Architectural, Sequence (Dynamic)
- [Technologies Used](#technologies-used)
- [Code examples (Code explained)](#code-examples-code-exaplained)
- [Challenges](#challenges)
- [Design patterns](#design-patterns)
- [Usage](#usage)

## Features

Coming up with a completely new chat application idea is challenging, and deciding which features to include or exclude is just as difficult. So, I chose to take inspiration from popular ones like Messenger and WhatsApp.

## Diagrams

- [ER diagram](./attachments/i9chat_ERD.png) - Created using the pgAdmin ERD tool.
- [Architectural Diagram](./attachments/i9chat_ARCHD.png) - A component-level diagram based on the c4 model. The API itself is a c4 "container" type interacting with other container types such as Databases, and Message queues. Created using Draw.io
- [Sequence Diagrams] - showing the flow of operations in the API for each endpoint accessed *(Coming soon...)*

## Technologies Used

### Core

- **Language/Runtime:** Go
- **Database System:** PostgreSQL.
  - Database Driver: pgx
- **Blob Storage:** Google Cloud Storage
- **Realtime Communication:** WebSockets
- **Messaging System:** Apache Kafka

#### PostgreSQL (Features used)

- Objects: Tables, Views,
  - **Stored Functions:** ,
  - and, **Types:**
- Cursor-based data fetching

### Frameworks & Packages

- API Framework
  - github.com/gofiber/fiber/v2
- Authentication
  - github.com/golang-jwt/jwt/v5
- E-Mailing
  - gopkg.in/gomail.v2
- Validation
  - github.com/go-ozzo/ozzo-validation/v4
- Database Driver
  - github.com/jackc/pgx/v5
- Blob storage
  - cloud.google.com/go/storage
- Realtime Communication
  - github.com/gofiber/contrib/websocket
- Event Streaming
  - kafka
- Security
  - golang.org/x/crypto/bcrypt
- Environment variables
  - github.com/joho/godotenv

### Tools

- Database Management
  - pgAdmin
  - psql
  - pg_dump, createdb, dropdp, pg_restore
- Functional Testing
  - Postman
- Version Control
  - Git & GitHub
  - Github Desktop
  - VSCode's "Source Control" Feature
- API Documentation
  - Markdown
- Workflow Speed-up
  - OpenAI's ChatGPT
  - GCP's integrated AI, Gemini
  - VSCode Extensions
  - Microsoft Bing Copilot
- Development
  - VSCode
  - Ubuntu Linux
  - Bash Script

## Code Examples (Code Exaplained)

The following are code examples with explanations of notable functionalities and solutions.

### Task: Creating a DM Chat

The API server handles this action on this WebSocket endpoint: ``

#### Sample WebSocket message

```js
{
  
}
```

#### Business Logic

The business logic is written in a `creatDMChat` service, called by the `createDMChat` controller/handler.

```go
```

...

## Challenges

## Design Patterns

### Singleton pattern

## Usage

Use the provided [API documention](./API%20doc.md).

> ⚠️ Note: The API documentation does not follow the OpenAPI specification, as it relies entirely on WebSocket communication for all endpoints. However, it is clearly written and easy to follow, similar to most standard documentation.

This API server is currently running remotely on Railway's cloud platform, but I plan to stop it very soon as I'll soon exhuast my trial plan.

However, you can run the project locally after cloning it to your computer by following these steps:

- Install and Setup Go
- Install and Setup PostgreSQL
- Install and Setup Apache Kafka
- Setup Google Cloud Storage on Google Cloud Platform
  - Create an API Key.
  - Create a bucket with desired name and make it publicly accessible.
- Setup Google SMTP with your gmail account and application password.
- Add a `.env` file cotaining these environment variables in the project's root folder.

  ```env
  PGDATABASE_URL="postgres://username:password@localhost:5432/i9chat_db"

  AUTH_JWT_SECRET=

  SIGNUP_SESSION_JWT_SECRET=

  MAILING_EMAIL=
  MAILING_PASSWORD=

  GCS_BUCKET=
  GCS_API_KEY=
  ```

- Get Kafka up and running
- Get PostgreSQL server up and running.
- Open a terminal in the project's root
  - Run `go mod tidy` to install dependencies.
  - Run `go install air` to install the live server
  - Run `air` in the project's root
  - Server will start on `http::localhost:8000`
