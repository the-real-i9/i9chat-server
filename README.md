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
- [API Graph Model Overview](#api-graph-model-overview)
- [API Documentation](#api-documentation)
- [API Diagrams](#api-diagrams)
- [Technical Highlights](#technical-notes)
  - [Why I switched from a relational database to a graph database.](#why-i-switched-from-a-relational-database-to-a-graph-database)
  - [Why I used the "chat entry" entity to generalize "message", "reaction", and "group activity" entities.](#why-i-used-the-chat-entry-entity-to-generalize-message-reaction-and-group-activity-entities)

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

  ![i9chat_api_arch](https://www.plantuml.com/plantuml/png/lLbjR-Eu4VwUNp7rXzO1v0JeBGeSMXItvsptjhqKkJXf-cr1KHEP6vdKaPGJzzSVXgGaKYzssitsKo9oI3upypp35_zWBDEssLKWlkTRPZ-pMdnccc_FVYSahve2uI_hTvVhTv_NQbShYTB-zUpC2biXT2rm9LYfsGuUbGxpmFkRJx11lKTzTlO7M2jfcP2ebwmM6In6HeiC5X_LulpiR7Cm5dUVFxpO9UwOgFnt1eiFz0N-qvkhiaBOND_TBCx3i2MlL5DihDAin6xacXeXj8vdIBe9GeVsAbyMpBAS6SnWSHN-fQ4tojXIu-QNdnVdYVtJtbYboVmVU5ZzYrKDkgaHuk6tM0WJZ_obmGQNk4TfDrOZSmxJryf_mZKQmqeXomY_cmMymSHiTPER9h_U4Dmclzmq-P5PWeLtF-GrZIRglU0u3gTlT_5SR-2Z0bu98aKbSitqWJhMIXfXh877Hd8hb3IEAcIVD_m9BUoOP2NkKDeCk99xr9PGglmBSWiE1qYsGmDMGQ6UPQLO0VUtFvFVjrXDGvramxr1Ct2JyQrpZFvs8ScWL2lj4ZNjWYKAU-10D04XeKsiW-IdxHbnYwmmbmzQs242WxpHmXu2SfF1OXEQsW0RprX2LJ5PhhuysmnohW-rJV2rKmrY_VU7Ep2Yb4AMbtlKub5mHhqPb2XHCnVkFwqAKGeBLZsXlEIgG7WMTWluKWljp04jlt1KQw4xZNRSk7OQZIZX09FDnVwF6JYuwj9MQKnP1zPXDIWBWkFqJmdtZ4gMZQJIh-uK4mGnHk6J2Sj_lo-AyoEq_9NKuzQdwGm8QHKCjKO9JbnFYrKW6LbD8Ks5xkJQhvPIGoEhhDX5-hiDBKCFpK7oYsjMnxxUD7abp3Q3JPCRhaMEvnayOsuKoUv841olAGYCMx57Q0ngyADC1YRCHWO-EKwOTPDVcYO7lcLIOkLdQXs0fYwO9MxN7XGDthc2v6NTZiew0XYtD_Z8zAQLU4AG-BoYFIHp_uK-_r4h4vgbiXtH7NCJCb5oHj6DYkQQLLKVmii-Wcc2A39KURMgAjJ69rz3r4HX_fl8KSF-kopKOBnX_2c9lyPV6pGMjamMDCe5puMzrkfHL9WKoKVXGwJi5hsPoO390agqJijk94N89BgXvfvA3cb2esErhjPgBmfqI8N68f0Yr2UAhLyq3WP1yXYZPl1d2vTYKSgcpa9jylOTO_eXi8oYWGObHyUNqcKZTUzQ5LYrphaFSkgVgDvzoM1zgBUeYHxgtPTLgLQ55djgU4snHybPRPgAQEp0h3JIHm7jCG4URZ_TVO1AbO87NsoZ9MWqJUNoudOAeY6hgYvFRGmlO6xfl15LaQvv_mbuwxY1HwrsSFlX_TMAPjLOuPv92suJWgKhN-U0qchHLQOHSrDoEFv7fbrF7EHPevcpmy0yJ1ETiWQRvC2mtwfdbXsNordJoIsajOutPqVHLE8Rr_P68hTRXEzMnc8TsE0dhxNYQ0mi9Ju3urmrqhOJXFq6qgsS2FSe9AlyF1fBOQpVkCzTPvZPe35K4lk2aeZ8M_Kn2Sx-xtKYBX41GbGD_kpHueimDWWR-PF9W3maCiTEWZkLE5z2eNCD_Gab6phf1DSUSgY_PiOyArsGdEayuXgvHjRNpH3j35XZrOhKB7r7VcXFKS3ScSeTbGAx9eWo32NHAx5EH3LDse7Xn5dnY7LwF9jXjv1ZTWlvrUpkTp2_o8XaGgOa4_88oQ93Oi8oKi-GQKGqQlUBH6IRvLf7YXspt1rDN4_M9xbDV1JdGTo6BWtJ4lKTptPaU2JRKJ1RzBTeqAQvl6aTSttVg69rqTDkaQvo3Te7nOih-IZqRbIykb5XXtTTWwbEbB0a87sW1y5ucuesG9PKDk_eW6vzPOiZ45WMbpRVWY-Y7jMkD0glgt4qTLpUXe4E39mdqqnKO_fEcypJVTh-H8sR7dIorD5f-BwcbuDK8cjAiUzu8ts4-mZii10C6ZyrQ-nMQV4VPbLQ2djS4Vm8qzlGrBsu1IWHHzl7YkdOJ4ABCZ7HV_pW5xbT28tSagDfEgvSExW8_DVf-EdkxgRB2H7-0VEDyczC8Ib9fZm87nmottKlPAeEQaelK3tgthaOyTiaAvunJTOKalOm7e_wTXSVlwwxVR_1_ITEm6klGLC8llesD4F7gNp72fZe7oVcn83ZbvQ90KUlBZDZJb_8HhYdRbiZe_cxMyJLKgkc7alz8pN_dvJk87YU_JwIkNipSLtzKM9E-vqNVVmY3zv6-T64SyBlOBJqYiPyEzcxLTzAzL4oNYlwYVM_MlFJj9n8Swo0SVSe8SVDJwftFEHuXHZRdnRx4FEKreSshv9w2xKFQVnctRVEgZphbKmsFpaH6shqxyzU-VTrgLc15t3jP6lQdvsuaXBzDT7u7mAMVjZv1JZGNSte7fGTSUXAStAAp17OW6MwH0iFct6f4RBTqlhAmWTgzaVnPnSyq5XN1yWPVsflhG66N_C2FbbWbL5XQV-OsbKC_y8-g2E3n-qu4tE7hZZjK-Dc2lYfOVC7_odhUPP7_A984HfPF2Zzv67ysBL1Q1p-x1ELEQYKB5UwaUuPPr1hk9AcsM6R-8uZom87lmlvQoc-86yiGeyAB3DFyHCtAKl8idSlL826LHrz2goeB1_Y1tDygL4QiQVx46qNfBxLZabHKqKcTYRRr8-2mZeHdVJ7fYllH_jp4t4_P6HYw43mNk_pvbCrPCBiI9f_G5aqk-et "i9chat_api_arch")

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



