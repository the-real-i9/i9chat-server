# i9chat (API Server)

[![Test i9chat](https://github.com/the-real-i9/i9chat-server/actions/workflows/test.yml/badge.svg?event=push)](https://github.com/the-real-i9/i9chat-server/actions/workflows/test.yml)

A Chat & Messaging API Server

## Intro

i9chat is a full-fledged chat and messaging API server built in Go It supports all major chat application application features with a scalable, production-grade arcitecture, serving as a foundation for building apps like WhatsApp and Discord clones.


## Technologies and Tools

<div style="display: flex; align-items: center;">
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/go-original-wordmark.svg" />
<img style="margin-right: 10px" alt="gofiber" width="40" src="./.attachments/tech-icons/gofiber.svg" />
<img style="margin-right: 10px" alt="redis" width="40" src="./.attachments/tech-icons/redis-original.svg" />
<img style="margin-right: 10px" alt="websocket" width="40" src="./.attachments/tech-icons/websocket.svg" />
<img style="margin-right: 10px" alt="neo4j" width="40" src="./.attachments/tech-icons/neo4j-original.svg" />
<img style="margin-right: 10px" alt="jwt" width="40" src="./.attachments/tech-icons/jwt.svg" />
<img style="margin-right: 10px" alt="googlecloud" width="40" src="./.attachments/tech-icons/googlecloud-original.svg" />
<img style="margin-right: 10px" alt="docker" width="40" src="./.attachments/tech-icons/docker-plain.svg" />
</div>


### Technologies
- **Go** - Programming Language
- **Fiber** - REST API Framework
- **Neo4j** - Graph DBMS
- **CypherQL** - Query Language for a Graph database
- **WebSockets** - Full-duplex, Bi-directional communication protocol
- **Redis Key/Value Store** (Cache)
- **Redis Streams**
- **Redis Pub/Sub**
- **Redis Queue** (via LPOP, RPUSH, RPOP, LPUSH)
- **Google Cloud Storage**

### Tools
- Docker
- Ubuntu Linux
- VSCode
- Git & GitHub Version Control
- GitHub Actions CI

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Technical Highlights](#technical-highlights)
- [API Documentation](#api-documentation)
- [Upcoming features](#upcoming-features)

## Features

The following are the features supported by this API. *Visit the API documentation for implementation guides.*

## Find Users

- Find a user by their username (exact matching only)
- Find users nearby (via geolocation coordinates)

## Chat & Messaging

Realtime chatting with users of the application, and in groups.

### Direct Chat

Start by finding a user by their username.

- Realtime user presence (online/offline) status and last seen.
- Supports various message types including:
  - Text
  - Voice
  - Image(s) with caption
  - Video(s) with caption
  - Audio
  - File attachments (Documents)
- React to Messages
- Reply to messages
- Delivered and Read receipts

### Group Chat

- Group creation
- Joining group
- Leaving group (you can't be re-added, unless you re-join)
- Total members count
- Online members count
- Group admin management
  - Add members
  - Remove members (they can't re-join, unless re-added)
  - Make member an admin
  - Remove member from admins

### Realtime Message Delivery

- Chat messages are delivered to target users in realtime.

### Real-time Updates

- Clients receive user "presence" and "last seen" updates (upon subscription)
- Real-time read receipts

## Technical Highlights

- 

## API Documentation

HTTP (REST) API: [Click Here](./docs/swagger.json)

WebSockets API: [Click Here](./docs/asyncapi.json)

