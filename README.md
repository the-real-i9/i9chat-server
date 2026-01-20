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
- [✨Technical Highlights✨](#technical-highlights)
- [API Graph Model Overview](#api-graph-model-overview)
- [API Documentation](#api-documentation)
- [API Diagrams](#api-diagrams)

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

## ✨Technical Highlights✨

- I switched from a relational database to a graph database when I saw that most hot queries involve more than two table JOINs, a point where SQL query performance starts to deplete, and a direct graph relationship shines.

- Chat history is served from a Redis Sorted Set, while I use Redis Stream’s stream message ID for ordering, so that received messages appear in the order they were delivered rather than in the order they were created, which is the way WhatsApp works.

- I use an abstract “ChatEntry” entity to represent an item in the chat history, which may be a message, a reaction, or a group activity. While each of these are an entity of their own, the “ChatEntry” generalization allows for a more efficient processing of a chat history READ request, where we won’t have to UNION multiple entity READs and re-order them; an approach that’s highly inefficient.

- All media processing and upload is offloaded to client-side. This eliminates a source of performance bottleneck, as a lot of user requests involving media processing will have to compete for server resources, an effect that will be felt even on user requests that don’t involve media processing unless additional infrastructure is added, which will incur unnecessary additional cost. Modern devices are powerful on their own when it comes to media processing.

## API Graph Model Overview

Read [Here](./docs/graph-model-overview.md)

## API Documentation

HTTP (REST) API: [Open Swagger JSON](./docs/swagger.json)

WebSockets API: [Open AsyncAPI JSON](./docs/asyncapi.json)


## API Diagrams

Architecture Diagrams: [See here](./diagrams/arch-diags.md)


