# Graph Model Overview

## Nodes (Entities)

- User
- DirectChat
- DirectChatEntry
- DirectMessage
- DirectMessageReaction
- Group
- GroupChat
- GroupChatEntry
- GroupMessage
- GroupMessageReaction

## Relationships
- `(:User)-[:HAS_CHAT]->(:DirectChat)-[:WITH_USER]->(:User)`
- `(:User)-[:SENDS_MESSAGE]->(:DirectMessage)`
- `(:User)-[:REACTS_TO_MESSAGE]->(:DirectMessage)`
- `(:DirectMessage)-[:IN_DIRECT_CHAT]->(:DirectChat)`
- `(:DirectMessageReaction)-[:IN_DIRECT_CHAT]->(:DirectChat)`
- `(:DirectMessage)-[:REPLIES_TO]->(:DirectMessage)`
- `(:DirectChatEntry)-[:IN_DIRECT_CHAT]->(:DirectChat)`

- `(:User)-[:IS_MEMBER_OF]->(:Group)`
- `(:User)-[:HAS_CHAT]->(:GroupChat)-[:WITH_GROUP]->(:Group)`
- `(:User)-[:SENDS_MESSAGE]->(:GroupMessage)`
- `(:User)-[:REACTS_TO_MESSAGE]->(:GroupMessage)`
- `(:GroupMessage)-[:IN_GROUP_CHAT]->(:GroupChat)`
- `(:GroupMessageReaction)-[:IN_GROUP_CHAT]->(:GroupChat)`
- `(:GroupMessage)-[:REPLIES_TO]->(:GroupMessage)`
- `(:GroupMessage)-[:DELIVERED_TO]->(:User)`
- `(:GroupMessage)-[:READ_BY]->(:User)`
- `(:GroupChatEntry)-[:IN_GROUP_CHAT]->(:GroupChat)`
