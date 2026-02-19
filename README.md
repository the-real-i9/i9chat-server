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
- **Fiber v3** - HTTP (REST) API Framework
- **Neo4j** - Graph DBMS
- **CypherQL** - Query Language for Neo4j
- **WebSocket** - Full-duplex, Bi-directional communications protocol. Realtime communication.
- **Redis Key/Value Store** - Cache. Fast data structures. Pagination. Aggregation.
- **Redis Streams** - Event-based messaging system. Background tasks queue.
- **Redis Pub/Sub** - PubSub pattern messaging system
- **Google Cloud Storage** - Cloud object storage
---
- **JWT** - User authentication. Token signing and verification.
- **MessagePack** - Object serializer and deserializer (major use)
- **JSON** - Object serializer and deserializer (minor use)

### Tools
- **Swagger** - HTTP API Documentation
- **AsyncAPI** - Websockets API Documention
- **Docker** - Container running Neo4j and Redis instances
- **Git & GitHub** - Repository & Version Control
- **GitHub Actions** - Continuous Integration. Unit & Integration Testing
- VSCode
- Ubuntu Linux

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [API Documentation](#api-documentation-)
- [API Diagrams](#api-diagrams-)
  - [Architecture Diagram](#architecture-diagram)
  - [Sequence Diagrams](#sequence-diagrams)
- [Articles](#articles-)

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


## API Documentation &#x1f4d6;

HTTP API (REST): [Here](./docs/swagger.json). Open in [Swagger Editor](editor.swagger.io)

WebSockets API: [Here](./docs/asyncapi.json) Open in [AsyncAPI Editor](studio.asyncapi.com).

API Graph Model Overview [Here](./docs/graph-model-overview.md)

## API Diagrams &#x1f3a8;

### Architecture Diagram
API (C4) Component Level Diagram: [Here](./arch.pu). (Open in [PlantUML Editor](editor.platuml.com))

<details>
  <summary>Show Diagram</summary>

  ![i9chat_api_arch](https://www.plantuml.com/plantuml/png/lLbjR-Is4VxkNp7T3yaQqDf0SoXGj2XoML-kryI5u_NL-NQWgB6MjrfI8Qcrjx--69AIIBsitPmld-mbX-GpCyypVD7tnZ9jctq5ugzyo-mdLejFJFFjsP-4v5LJ8Fnz_UPo_URJMkrh9L7QVvoTMM4hXAu5hWGhDTl3Wz9X7dXxym4sg0-epyw-XRMIbWc9UibgaS5YPBJ8OF5UBSxFpZP7Ot7_wTsJNUAUYSh_pc3nZdw1_-dDLLaXRAxlRXRdOTYILwefDbPfLc8tSasD45h7CoJT1A53UvKl2sPPpWnciBWA_zBG6sLigN7poy-ByyH-Tp1MQJB_2O-h_x2gGJSrGZpy5WjXuf6_DDZW4WyexSPgPCvX-hNoF-4QZM6ba6M4tyq2tc1Yjhh9JPDVRmXkarzkclp4BC72kn_ocaOJzK5m78VJjxjuhZVmNW6l196Y4hbc-aWTQoMDCDR0GoEv5KeQHnMopnk-Gmjx9bc9UvGs0wxa0RKbb2h_ZDo2mm6IxT60LL2eHrafLi37sv_9xrkifg5Eik6ZGJDma_6jSup-TY79e5GhxH8rxO8b2dlWG3G18Q5Dh8Faf-qPSOkiC9VtMjWX0eEyqS8U0tAJmM8Jcje06yzOGbKnMQu-FzeCSgwFjKtmjLCDOlsl-pimefH2bfS7rE91S4Qz6PGeKJCNxh-i2b4A2rRjK5voLI0y2hi5VAg5TkQ0bbyughNGdKOxRbmxZKQKS839vkBmvmmSN7LfgpIch0Fhi1gK1S5n-gU4UqOboqPIwLTtYWc26AFmoOJbhs-Bulm4BNyZzRZrQJe3Gbe5GwrHWXFNqs8LI4PMKqXJeJkvzgkbrD38AYlsaVvkGylGGtEK_EAQrR6lDqrU2RFDODFaXckHuta6ZvWRHR8x4GJ7Qme2urOS41g36lmeau49iv61JuuJPjtabwR9WM-Pb5ZvcLe7eAaBPedRjGT5W_Uk87bPjwEoZW36xGq-Chrf9NO84XzNj8Taxh_Gvt_KoeHcgMn7TCVSX4oKd55qewAvPbNLX_2oZs2Q88eCLLwjgWgrySdN43LHc7yKEMeuV9U56emtZE-I-6lyhK5ZOSjaGQDSy5pOQwqUH8L9aNmGFaJARj6RcGoOBA14wxJiHbA4JA8RQkwfv90cD3fLwsgj3g90XrHeB08fGdsYsFf5us0G98ynQWP_jd0f5gLiwYpK9c_VCQOV0iie6cXGSdHyALqsKlUkLO5LwvpxGKxz6zMRpnciZ_KMDT53lVcyAjMgqE903IyfzYWvgqrJ4OqTc9L6-b50Uqo0-ziFT--WKgNWmHVRQ0aQJLEvlBYTWcZ8gghBKnl32vXREcz4LQHhVlm0l7NSm8DMUxXzz_PgHRDghF30f0Mt2S7IbQzpG6crw2hJYBafEHp_8zEk9mxohD7CsM5W7gQ9Jbc3JN9Ws6_LCykEo-MiwUGMqbh7cxEZQ2hnZMjx9P7RBS9tAsEn3cpmazTQSJG6bX8VWN6k6cdR2S9-0-bMJeHx499LVXwDfJ3MRzpdhZFCRD0OAebzGKb4vAtwc0HdVt-TY4j4W11LWxzvjFWaZ0t2Hhup6P27HEROIN2d4kTBA7IkeJ-XP4CddU3QGmxrryoOHwKBaZETHrmZryZwkXcYdG5hh5gHcgNlo8_jAGgOErEver9WrmHHXg4aUYNMYQYcIJimd3Whdh1EZsSpx1Pop6uXlvZT_Gxc5ncHJCYKP4AU85bqI4nOHkePCes8Hkr-aOXiirphIB5dbhkZYUl9-YItYO_YF8ZRqALXMgA-qzcE34zaEmfcY_uM3TeqbpUjOwxlArMiBdhQBT9Lha7xe7Xn9H-4te-ALpSgxF2kQp3LCoKi2KWVw44mNgUY3P0bbSqxEg1RNzdY20IMnQNDj-2Jg4UrAut2oseSJHsNjs6W0mFdoJGJrPY-qoRpT9-sVwR6JGzwjjJHQVXZJIy7gKJMb6BVyKhw2FSHs6CX63H-QjROhTBYlyogj1Jsk2Bu4QRNeQbxSGjG8ewsZnNJiPc45cLYel_qmIzokn0QkIL7qtHSkNPm4VY_q_7FkxkRBYT4-7lCDyg_COMa99do87nmo7pVlP2fEgWflK3rg7ldOiIlawnunZHPKqZQmtWywkjTVFowxlRz2lUVE04llGPD8VZYsz0C7QVo7IjWed-Sc6S6dBwqJ0muUNMP6VFyXMo4U-gsDJAQlxj5N2srQkgnrDzJyn-aT0V1y-oRIERkpSHrzKU9EUrtNlJnYpnu7UL74yu9luFHqYiQy_LaxrPzBTL7oNYfw2VM_sBDJzDo8isn0iRTeuISDdzMlUCXfol4sFvviGynJsbzQFCYgRTG-v36RzPzwwpAirx9PFEMSwpf_YVT2GgbfQS0BC93xhU6pCFJD4b68W_l0FJPnRYxIlWarVUwYb0hOqHUONXdNbcr2eRfY_C5E6-xOjEzH3l6JSwVGO7qC47NrkvncewjQC1zO07BT8dHiL78TYly9bP7_6ggT7XATOFQ1t9_FYRLFRfiWCQwEhOtu16oYy7BTvockHDPp8cfSJFL_RbXyxU2AUjv2KJaondV4HfPt2kzyp1-wDeWD0w_2KKr42ebovLkf7lZ6HGYhgHfzWZ6Vs7gUB0iSF3HoDzPyGbvOn5wL62PsSNltoGx8SlUlL0U6bPrsrY15TNiO_oQZayrIYCET5cYlOIakTepAMegryJmP0_xGL1O9wAJVecwydwqtwA8to4Z4qC7rS0NWtdpgG8pONOYpT-ZB9fzzJy0 "i9chat_api_arch")

</details>

### Sequence Diagrams

API sequence diagrams: [Here](./diagrams/sequence-diagrams.md)

## Articles &#x1f4f0;
*Coming Soon...*

<!-- ## Technical Highlights Notes

### Why I switched from a relational database to a graph database

I initially wrote the whole API database logic using PostgreSQL.

But, after being introduced to the world of graph databases, I recalled that most of my SQL database logic, especially those of group management, involved a combination of several READ and WRITE queries. Even after employing stored functions, each statement still performs a full index scan.

A Neo4j query, on the other hand, is procedural, has the ability to retain pointer to nodes through intermediate clauses, and performs its scan within relevant branches of the graph. These proved Neo4j a more efficient alternative.

### Why I used the "chat entry" entity to generalize "message", "reaction", and "group activity" entities.

A chat history collection is a union of **message**, **reaction**, and a **group activity** (for group chat) entities each with different structure.

To fetch a chat's history (direct or group), the predictable approach is to UNION these entities, and re-order them using a common `date_created` property, this is a very expensive approach.

While each of these entities are treated separately for other purposes, the **"direct/group chat entry"** generalization allows us to treat these entites as the same only for the purpose of fetching chat history, more efficiently.

### Why I enforce value limits where possible.

...
 -->
<!-- - I switched from a relational database to a graph database because, most of my SQL database logic, especially those of group management, involved a combination of several READ and WRITE queries. Even after employing stored functions, each statement still performs a full index scan. A Neo4j query, on the other hand, is procedural, has the ability to retain pointer to nodes through intermediate clauses, and performs its scan within relevant branches of the graph. These proved Neo4j a more efficient alternative.

- Chat history is served from a Redis Sorted Set, while I use Redis Stream’s stream message ID for ordering (ZSET score), so that received messages appear in the order they were delivered rather than in the order they were created—the way WhatsApp works.

- I use an abstract “chat entry” entity to represent an item in the chat history, which may be a message, a reaction, or a group activity. While each of these are an entity on their own, the “chat entry” generalization allows for a more efficient processing of READ requests for a chat’s history, where we won’t have to UNION multiple entity READs and finally re-order them, a highly inefficient approach.

- All media processing and upload is offloaded to client-side. This eliminates a source of performance bottleneck, as many simultaneous user requests involving media processing will have to compete for server resources, an effect that will be felt even on user requests that don’t involve media processing unless additional infrastructure is added, which will incur unnecessary additional cost. The average modern device today has powerful media processing capabilities. -->



