# i9chat (API Server)

[![Test i9chat](https://github.com/the-real-i9/i9chat-server/actions/workflows/test.yml/badge.svg)](https://github.com/the-real-i9/i9chat-server/actions/workflows/test.yml)

Build a robust Chat Application

## Intro

i9chat is a REST API server for a Chat Application, built using Go and Neo4j. It supports major chat application features.

## Technologies

<div style="display: flex;">
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/go-original-wordmark.svg" />
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/gofiber.svg" />
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/websocket.svg" />
<img style="margin-right: 10px" alt="neo4j" width="70" src="./.attachments/tech-icons/neo4j-original.svg" />
<img style="margin-right: 10px" alt="nodejs" width="40" src="./.attachments/tech-icons/apachekafka-original.svg" />
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/jwt.svg" />
<img style="margin-right: 10px" alt="nodejs" width="40" src="./.attachments/tech-icons/googlecloud-original.svg" />
<img style="margin-right: 10px" alt="postgresql" width="40" src="./.attachments/tech-icons/postgresql-original.svg" /> ❌ (old)
</div>

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [API Documentation](./API%20doc.md)

## Features

The following is a summary of the features supported by this API. Visit the API documentation to see the full features and their implementation details.

### DM Chat

- Message of different types including text, voice, video, audio, photo, and file attachments.
- React to Messages
- Unsend Messages
- Delete Messages

### Group Chat

- Everything in DM Chat
- Start a Group, Add members, Make Admins etc.
- Join Group, Leave Group etc.

### Make Friends

- You can search for friends by their email, or username.
- Send friend request
- Friend Suggestions

### Moments (Upcoming)

- A say for Statuses on WhatsApp, and Stories on Messenger

## API Documentation

For all **REST request/response Communication**: [Click Here](./.apidoc/restapi.md)

For all **WebSocket Real-time Communication**: [Click Here](./.apidoc/websocketsapi.md)
