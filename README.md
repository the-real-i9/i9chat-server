# API Documentation

**i9chat** is a chat application server API built entirely with WebSocket technology. This document provides a conprehensive usage instructions for client applications.

## Important Notes

Before we dive in, separated by horizontal lines are informations we need to keep in mind.

**⚠ Note:** All connections must be explicitly closed because they remain open to allow retries without needing to establish a new connection in case of errors.

---
"Sent" and "Received" data implies, "the data to send to the server" and "the expected response", respectively.

---
The are two **formats of received data** :

- The first format indicates a **success response**:

  ```json
    {
      "statusCode": 200,
      "body": {}
    }
  ```

- The second format indicates an **error response** (if an error has occured):

  ```txt
    {
      "statusCode": 4xx,
      "error": "reason for error"
    }
  ```

---
The structure of a message's content, sent and received during interaction, is fully described [here](#message-content).

## API Endpoints

All communication is implemented with WebSocket, therefore, all URLs are of the `ws://` protocol. Futhermore, all sent/received data are of the JSON data type.

The endpoint usage instructions follow a defined pattern:

1. A description of the endpoint and its usage.
2. The endpoint URL.
3. The sent header, if any. Most likely, it'd be the `Authorization` header.
4. A description of each type of sent data and its corresponding received data, accompanied by examples.
5. Final instructions on the endpoint, if any.

The application's endpoints are organized into three categories: user authentication, user actions, and chat actions.

### User Authentication

This category includes endpoints handling all aspects of user authentication.

- Signup
  - [Request new account](#sign-up-request-new-account---step-1)
  - [Verify email](#sign-up-verify-email---step-2)
  - [Register user](#sign-up-register-user---step-3)
- [Signin](#sign-in)

### User Actions

This category includes endpoints handling user actions.

- [Change profile picture](#change-profile-picture)
- [Update client's geolocation](#update-clients-geolocation)
- [Switch client's presence](#switch-clients-presence)
- [Get all users](#get-all-users)
- [Search user](#search-user)
- [Find nearby users](#find-nearby-users)
- [Get my chats](#get-my-chats)
- [Open DM Chat Stream](#open-dm-chat-stream)
- [Open Group Chat Stream](#open-group-chat-stream)

### Chat Actions

This category includes endpoints handling chat session operations.

- [Get DM chat history](#get-dm-chat-history)
- [Get Group chat history](#get-group-chat-history)
- [Open DM Messaging Stream](#open-dm-messaging-stream)
- [Open Group Messaging Stream](#open-group-messaging-stream)

## Endpoints and Usage

Creating an account (Sign up) is divided into three steps. Each step is uniquely identified by a URL and *must be accessed in order*.

### Sign up: Request New Account - (Step 1)

Here, the client submits their email and a verification code is sent to it.

**URL:** `ws://localhost:8000/api/auth/signup/request_new_account`

**Sent Data:**

```json
{
  "email": "example_user@gmail.com"
}
```

**Received Data:** (Success)

The `signupSessionJwt` is a session token to establish a session for the signup process, just like a session cookie.

```json
{
  "statusCode": 200,
  "body": {
    "msg": "A 6-digit verification code has been sent to example_user@gmail.com",
    "signupSessionJwt": "${jwt}"
  }
}
```

**Received Data:** (Error)

```json
{
  "statusCode": 422,
  "error": "An account with this email already exists"
}
```

Finally, close the open connection and navigate the user to the email verification page.

### Sign up: Verify Email - (Step 2)

Here, the client provides the verification code sent to the email in the previous step. Any attempt to access these steps out of order will result in an error.

**URL:** `ws://localhost:8000/api/auth/signup/verify_email`

**Sent Header:** `Authorization: Bearer ${signupSessionJwt}`

**Sent Data:**

```json
{
  "code": 123456
}
```

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body": {
    "msg": "Your email example_user@gmail.com has been verified"
  }
}
```

**Received Data:** (Error)

```json
{
  "statusCode": 422,
  "error": "Incorrect verification code. Check your email or re-submit your email",
  "error": "Verification code has expired. Re-submit your email"
}
// Note: The duplicated `error` keys here are just to provide the two possible errors you can get.
```

### Sign up: Register User - (Step 3)

Here, the client provides the user's account information to be used for registration.

**URL:** `ws://localhost:8000/api/auth/signup/register_user`

**Sent Header:** `Authorization: Bearer ${signupSessionJwt}`

**Sent Data:**

Note: `geolocation` data should not be provided explicitly by the user, rather it should be provided somehow by the client application.

```json
{
  "username": "abc",
  "password": "blablabla",
  "geolocation": "2, 5, 5"
}
```

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body": {
    "msg": "Signup success!",
    "user": {
      "id": 1,
      "username": "abc"
    },
    "authJwt": "${authJwt}"
  }
}
```

**Received Data:** (Error)

```json
{
  "statusCode": 422,
  "error": "Username unavailable", 
}
```

### Sign in

**URL:** `ws://localhost:8000/api/auth/signin`

**Sent Data:**

```json
{
  "emailOrUsername": "abc",
  "password": "blablabla",
}
```

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body": {
    "msg": "Signin success!",
    "user": {
      "id": 1,
      "username": "abc"
    },
    "authJwt": "${authJwt}"
  }
}
```

**Received Data:** (Error)

```json
{
  "statusCode": 422,
  "error": "Incorrect username or password"
}
```

### Change profile picture

**URL:** `ws://localhost:8000/api/app/user/change_profile_picture`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Sent Data:**

The value `pictureData` is of a binary data representation, a `uint8` array, precisely. Most programming languages have a data type of this representation.

```json
{
  "pictureData": [97, 98, 99, 100, 100, 101, 102]
}
```

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body": {
    "msg": "Operation successful",
  }
}
```

**Received Data:** (Error)

```json
{
  "statusCode": 422,
  "error": "Operation failed: ...", 
}
```

### Update Client's Geolocation

**URL:** `ws://localhost:8000/api/app/user/update_my_geolocation`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Sent Data:**

Clients should be able to update their current location if, perhaps, they move to a new place and want to be seen nearby.

Format: `latitude, longitude, center` or `(latitude, longitude), center`

```json
{
  "newGeolocation": "5, 2, 4"
}
```

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body":  {
    "msg": "Operation successful"
  }
}
```

### Switch Client's Presence

The client application should detect and update the client's presence on the server.

**URL:** `ws://localhost:8000/api/app/user/switch_my_presence`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Sent Data:**

The value of `presence` is either `online` or `offline`.

If `presence`'s value is `offline`, then the value of `lastSeen` should be the time the client went offline, else the value of `lastSeen` should be `null`.

```json
{
  "presence": "offline",
  "lastSeen": "${dateString}"
}

{
  "presence": "online",
  "lastSeen": null
}
```

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body":  {
    "msg": "Operation successful"
  }
}
```

### Get all users

For some reason you might want to get all users, perhaps, to start a new dm chat, create a group or add participants.

**URL:** `ws://localhost:8000/api/app/user/all_users`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body":  [
    {
      "type": "dm",
      "chat_id": 1,
      "partner": {
        "user_id": 4,
        "username": "risotto",
        "profile_pic_url": "someurl.jpg"
      },
      "updated_at": "${some date}"
    },
    {
      "type": "group",
      "chat_id": 2,
      "name": "Class of 2018",
      "description": "Still wondering what we are...",
      "picture_url": "someurl.jpg",
      "updated_at": "${some date}"
    }
  ],
}
```

### Search User

For some reason you might want to search user, perhaps, to start a new dm chat, narrow the search result when creating a group or adding participants.

**URL:** `ws://localhost:8000/api/app/user/search_user`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Sent Data:**

```json
{
  "query": "da"
}
```

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body":  [
    {
      "user_id": 4,
      "username": "david",
      "profile_pic_url": "someurl.jpg"
    },
    {
      "user_id": 5,
      "username": "daemon",
      "profile_pic_url": "someurl.jpg"
    }
  ],
}
```

### Find Nearby Users

**URL:** `ws://localhost:8000/api/app/user/find_nearby_users`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Sent Data:**

The client application should provide a location value at present time: `latitude, longitude, center` or `(latitude, longitude), center`

```json
{
  "liveLocation": "5, 2, 4"
}
```

**Received Data:** (Success)

```json
{
  "statusCode": 200,
  "body":  [
    {
      "user_id": 4,
      "username": "david",
      "profile_pic_url": "someurl.jpg"
    },
    {
      "user_id": 5,
      "username": "daemon",
      "profile_pic_url": "someurl.jpg"
    }
  ],
}
```

### Get my chats

Access this endpoint provided the client's chat list is not previously cached. After which you must access the endpoints: `.../open_dm_chat_stream` and `.../open_group_chat_stream` below.

**URL:** `ws://localhost:8000/api/app/user/my_chats`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Received Data:** (Success)

A list consisting all user's DM and Group chats in descending order by time updated. **It is recommended that you cache this result.**

```json
{
  "statusCode": 200,
  "body": [
    {
      "type": "dm",
      "chat_id": 1,
      "partner": {
        "user_id": 4,
        "username": "risotto",
        "profile_pic_url": "someurl.jpg"
      },
      "updated_at": "${some date}"
    },
    {
      "type": "group",
      "chat_id": 2,
      "name": "Class of 2018",
      "description": "Still wondering what we are...",
      "picture_url": "someurl.jpg",
      "updated_at": "${some date}"
    }
  ],
}
```

### Open DM Chat Stream

This stream (endpoint) should open (be accessed) as soon as the client is online, and remain open until the client goes offline, after which it is closed. This, basically, is how the client lets the server know they are online or offline.

> Note! The structure of a chat message is defined in the "Send message" endpoint section.

**URL:** `ws://localhost:8000/api/app/user/open_dm_chat_stream`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Received Events:** (*Study carefully!!*)

This stream receives "events" sent from the server.

At every opening of the stream (when user comes online), all events pending receipt and associated data are first streamed to the client.

The events, associated data, recommended listeners, and additional information (if any) are discussed below:

- The `new chat` event:

  - Received when another user initiates a new DM chat with the client. Here's the data's structure:

    ```json
    {
      "event": "new chat",
      "data": {
        "type": "dm",
        "dm_chat_id": 2,
        "partner": {
          "id": 2,
          "username": "samuel",
          "profile_picture_url": "someurl.jpg",
          "presence": "online",
          "last_seen": null
        },
        "init_msg": {
          "id": 5,
          "content": {
            "type": "text",
            "props": {
              "textContent": "Hi! How're you?"
            } 
          }
        }
      }
    }
    ```

  - The UI that lists the client's chats should listen for this event, so as to stack this new DM chat on top of the list.
  - The application should set the rendered chat snippet's "unread messages count" to `1`, to reflect the new chat's associated initial message.
  - If the application includes the "latest message" on the chat snippet UI, the message content associated with `init_msg` should be used.
  - The application should send acknowledgement for the initial message attached. How to achieve this is described later in this section.

- The `new message` event:

  - Received when the client gets a new message in an existing DM chat.

    ```json
      {
        "event": "new message",
        "data": {
          "msg_id": 2,
          "dm_chat_id": 4,
          "sender": {
            "id": 5,
            "username": "samuel",
            "profile_pic_url": "someimage.jpg"
          },
          "content": {
            "type": "text",
            "props": {
              "textContent": "Hi! How're you?"
            } 
          },
          "reactions": []
        }
      }  
    ```

  - The UI that lists the client's chats should listen for this event, so as to find the target DM chat snippet,

    - update its "unread messages count" `+1`,
    - update its "lastest message" with the content of the new message,
    - update its "last updated time" to the time at which the message was received.

    **But,** if the target DM chat is open, this new message is appended to the chat history list, rather than updating the chat snippet's "unread messages count" `+1`. Other chat snippet updates (listed above) are done.
  - The application should send acknowledgement for this new message. How to achieve this is described later in this section.

**Sent Data:**

This stream allows sending three types of data, determined by the `action` to execute.

Each type of `action` has its associated data.

- First, to create a new DM chat, send:

    ```json
    {
      "action": "create new chat",
      "data": {
        "partnerId": 2,
        "initMsg": {
          "type": "text",
          "props": {
            "textContent": "Hi! How're you?"
          }
        },
        "createdAt": "${dateTimeString}"
      }
    }
    ```

- Second, to send acknowledgement for received messages

    The value of `status` can either be `delivered` or `seen`. The property `at` indicates the time of delivery.

    ```json
    {
      "action": "acknowledge message",
      "data": {
        "status": "delivered",
        "msgId": 2,
        "dmChatId": 4,
        "senderId": 2,
        "at": "${dateTimeString}",
      }
    }
    ```

- The third is like the second, with an extra feature that lets you acknowledge received messages (with the same status) in batch.

    ```json
    {
      "action": "batch acknowledge messages",
      "data": {
        "status": "delivered",
        "msgAckDatas": [
          {
            "msgId": 2,
            "dmChatId": 4,
            "senderId": 2,
            "at": "${dateTimeString}",
          },
          {
            "msgId": 3,
            "dmChatId": 4,
            "senderId": 3,
            "at": "${dateTimeString}",
          }
        ]
      }
    }
    ```

**Received Data:**

The only data received is the servers's response to the client's `create new chat` send action.

> Recall that, the partner with which the client has initiated this new chat will receive thier own data in a `new chat` event (as described above). This happens immediately provided the partner is online, else, the event is queued as part of the "dm chat events pending receipt" for the user, which will be streamed to them as soon as they come online.

The client is expected to perform an "optimistic UI" rendering with the data used to create the DM chat, hence, only the needful data is returned.

```json
{
  "new_dm_chat_id": 4,
  "init_msg_id": 5
}
```

### Open Group Chat Stream

This stream (endpoint) should open (be accessed) as soon as the client is online, and remain open until the client goes offline, after which it is closed. This, basically, is how the client lets the server know they are online or offline.

**URL:** `ws://localhost:8000/api/app/user/open_group_chat_stream`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Received Events:** (*Study carefully!!*)

This stream receives "events" sent from the server.

At every opening of the stream (when user comes online), all events pending receipt and associated data are first streamed to the client.

The events, associated data, recommended listeners, and additional information (if any) are discussed below:

- The `new chat` event:

  - Received when the client is added to a new or existing Group chat. Other participants will also receive this event (excluding the creator of the group chat).

    ```json
    {
      "event": "new chat",
      "data": {
        "type": "group",
        "group_chat_id": 2,
        "name": "Class of 2018",
        "picture_url": "somegrouppic.jpg"
      }
    }
    ```

  - The UI that lists the client's chats should listen for this event, so as to stack this new group chat on top of the list.
  - If the application includes the "latest activity" on the chat snippet UI, it should read *"You were added"*.

- The `new message` event:

  - Received when a new message is sent to one of your group chats. Other group members, excluding the sender, will also receive this event.

    ```json
      {
        "event": "new message",
        "data": {
          "msg_id": 2,
          "group_chat_id": 4,
          "sender": {
            "id": 5,
            "username": "samuel",
            "profile_pic_url": "someimage.jpg"
          },
          "content": {
            "type": "text",
            "props": {
              "textContent": "Hi! How're you?"
            } 
          },
          "reactions": []
        }
      }  
    ```

  - The UI that lists the client's chats should listen for this event, so as to find the target group chat snippet,

    - update its "unread messages count" `+1`,
    - update its "lastest message" with the content of the new message,
    - update its "last updated time" to the time at which the message was received.

    **But,** if the target group chat is open, this new message is appended to the chat history list, rather than updating the chat snippet's "unread messages count" `+1`. Other chat snippet updates (listed above) are done.
  - The application should send acknowledgement for this new message. How to achieve this is described later in this section.

**Sent Data:**

This stream allows sending two types of data, determined by the `action` to execute.

Each type of `action` has its associated data.

- First, to create a new Group chat, send:

    The `pictureData` is a binary data represented as a `uint8` integer array, as described earlier. The data will be uploaded to a cloud storage and replaced with the result URL.

    The `initUsers` is a 2D-array containing the `[id, username]` of the initial participants with which the group will be created.

    ```json
    {
      "action": "create new chat",
      "data": {
        "name": "Class of 2018",
        "description": "We don't suck, you see!",
        "pictureData": [97, 98, 99, 100, 101, 102],
        "initUsers": [
          [2, "samuel"],
          [4, "david"],
          [5, "dave"],
        ]
      }
    }
    ```

- Second, to send acknowledgement for received messages

    The value of `status` can either be `delivered` or `seen`. The property `at` indicates the time of delivery.

    Seeing we're dealing with a group, it is more efficient to acknowledge messages in batch rather than singly. Also, observe that you can only acknowledge messages for a single group at a time.

    ```json
    {
      "action": "acknowledge messages",
      "data": {
        "status": "seen",
        "groupChatId": 4,
        "msgAckDatas": [
          {
            "msgId": 2,
            "at": "${dateTimeString}",
          },
          {
            "msgId": 3,
            "at": "${dateTimeString}",
          }
        ]
      }
    }
    ```

**Received Data:**

The only data received is the servers's response to the client's `create new chat` send action.

> Recall that, all participants included in the creation of the group chat will receive their own data in a `new chat` event (as described above). This happens immediately provided the participant is online, else, the event is queued as part of the "group chat events pending receipt" for the user, which will be streamed to them as soon as they come online.

The client is expected to perform an "optimistic UI" rendering with data it uses to create the group chat, hence, only the needful data is returned.

```json
{
  "new_group_chat_id": 4
}
```

### Get DM Chat History

**URL:** `ws://localhost:8000/api/app/dm_chat/chat_history`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Sent Data:**

The `offset` is how we implement infinite scrolling or pagination. An `offset = 0` returns the latest messages in the history. The default `limit = 50`; you won't need to change it, therefore, it is not allowed to be set.

I don't think I have to explain how infinite scrolling works, nevertheless, just do `offset = numberOfRequests * limit` on each request.

```json
{
  "dmChatId": 2,
  "offset": 0
}
```

**Received Data:**

DM chat history consists of messages only, unlike a group, it contains no activity. Therefore, all data should be treated as a message data, and be used to render the message snippet.

You should know that, if the client is the sender (`clientId = sender.id`), you should render its message snippet different from that of its partner. Basically, you don't include read receipts on partner's message snippet, but you do on client's message snippet.

```json
[
  {
    "id": 2,
    "sender": {
      "id": 4,
      "username": "samuel",
      "profile_picture_url": "someurl.jpg"
    },
    "content": {
      "type": "text",
      "props": {
        "textContent": "Hey bro!"
      }
    },
    "delivery_status": "seen",
    "created_at": "${dateTimeString}", 
    "edited": false,
    "edited_at": "${dateTimeString}", 
    "reactions": [
      {
        "reaction": "${reactionCodePoint}",
        "reactor": {
          "id": 4,
          "username": "samuel",
          "profile_picture_url": "someurl.jpg"
        }
      }
    ]
  }
]
```

### Get Group Chat History

**URL:** `ws://localhost:8000/api/app/dm_chat/chat_history`

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Sent Data:**

The `offset` is how we implement infinite scrolling or pagination. An `offset = 0` returns the latest messages in the history. The default `limit = 50`; you won't need to change it, therefore, it is not allowed to be set.

I don't think I have to explain how infinite scrolling works, nevertheless, just do `offset = numberOfRequests * limit` on each request.

```json
{
  "groupChatId": 2,
  "offset": 0
}
```

**Received Data:**

Group chat history consists of messages and activity. The type of data (wheter a "message" or an "activity") is determined by the data's `type` property. A data of a certain type has its associated properties to be used to render the corresponding snippet.

You should know that, if the client is the sender (`clientId = sender.id`), you should render its message snippet different from that of other group members. Basically, you don't include read receipts on other group members' message snippet, but you do on client's message snippet.

In "activity" history type, the structure of `activity_info` is based on `activity_type`. These data should be used to render the activity message appropriately. The sample data below represents an activity that should reads *"dave joined"*.

```json
[
  {
    "type": "message",
    "id": 2,
    "sender": {
      "id": 4,
      "username": "samuel",
      "profile_picture_url": "someurl.jpg"
    },
    "content": {
      "type": "text",
      "props": {
        "textContent": "Hey bro!"
      }
    },
    "delivery_status": "seen",
    "created_at": "${dateTimeString}", 
    "edited": false,
    "edited_at": "${dateTimeString}", 
    "reactions": [
      {
        "reaction": "${reactionCodePoint}",
        "reactor": {
          "id": 4,
          "username": "samuel",
          "profile_picture_url": "someurl.jpg"
        }
      }
    ]
  },
  {
    "type": "activity",
    "activity_type": "user_joined",
    "activity_info": {
      "username": "dave"
    }
  }
]
```

### Open DM Messaging Stream

This stream (endpoint) is where the client *sends messages* and *receives delivery acknowledgements* for a particular DM chat (identified by the parameter, `:dm_chat_id`, at the last segment of the URL).

This stream (endpoint) should be opened (accessed) as soon as the client enters a DM chat session, and closed when the client leaves this session off to another, whose stream also opens for messaging.

**URL:** `ws://localhost:8000/api/app/dm_chat/open_dm_messaging_stream/:dm_chat_id`.

*`:dm_chat_id` must be replaced with the target DM chat's unique id in the URL (`..._stream/5`, where `5` is the target DM chat's unique id).*

**Sent Header:** `Authorization: Bearer ${authJwt}`

**Received Event:** (*Study carefully!!*)

This stream receives only one type of event sent from the server, *whenever the client receives a message delivery acknowledgement for one of its messages.*

> More events may be supported in the near future, a "message edit" event, for example.

At every opening of the stream (when the client enters a DM chat session), all events pending receipt and associated data are first streamed to the client.

Below is the event, associated data, and recommended listeners:

```json
{
  "event": "delivery status update",
  "data": {
    "msgId": 5,
    "status": "seen"
  }
}
```

The DM chat messaging interface (where message is sent) should listen for this event, find the target message through `msgId`, and update the read receipt accordingly.

**Sent Data:**

The data sent through this stream is *a message*.

> Note! Messages are not received on this stream, the receipt of messages has been [discussed](#open-dm-chat-stream).

The value of `at` specifies the time the message was created.

```json
{
  "msg": {
    "type": "text",
    "props": {
      "textContent": "Hi! How're you doing?"
    }
  },
  "at": "${datetimeString}"
}
```

Below, the structure(s) of `msg` content is described in detail.

#### Message Content

A message can be one of the following types: `text`, `voice`, `image`, `audio`, `video`, `file`.

The `props` property of a message data, holds the properties associated with a message of a particular type.

Messages that include the binary data for their type, have a `uint8` integer array representation of the data. This binary data representation will be uploaded to a cloud storage, and the `{type}Data` property is replaced with `{type}Url` which holds the resulting URL from the upload.

The data format for each type of message is described below:

- Type: `text`

  ```json
  {
    "type": "text",
    "props": {
      "textContent": "Hi! How're you doing?"
    }
  }
  ```

- Type: `voice`

  The value of `duration` is specified in seconds. Below is a `200` duration sec.

  ```json
  {
    "type": "voice",
    "props": {
      "voiceData": [97, 98, 99, 100, 101, 102, 103, 104, 105],
      "duration": 200
    }
  }
  ```

- Type: `image`

  The value of `size` is specified in bytes.

  ```json
  {
    "type": "image",
    "props": {
      "imageData": [97, 98, 99, 100, 101, 102, 103, 104, 105],
      "mimeType": "image/png",
      "caption": "This is an image",
      "size": 4076
    }
  }
  ```

- Type: `audio`

  The value of `size` is specified in bytes.

  ```json
  {
    "type": "audio",
    "props": {
      "auioData": [97, 98, 99, 100, 101, 102, 103, 104, 105],
      "mimeType": "audio/mp3",
      "caption": "Enjoy the music",
      "size": 4076
    }
  }
  ```

- Type: `video`

  The value of `size` is specified in bytes.

  ```json
  {
    "type": "video",
    "props": {
      "videoData": [97, 98, 99, 100, 101, 102, 103, 104, 105],
      "mimeType": "video/mp4",
      "caption": "Blockbuster baby!",
      "size": 4076
    }
  }
  ```

- Type: `file`

  The value of `size` is specified in bytes.

  ```json
  {
    "type": "file",
    "props": {
      "fileData": [97, 98, 99, 100, 101, 102, 103, 104, 105],
      "mimeType": "application/octet-stream",
      "caption": "This is a text file",
      "size": 4076,
      "extension": "txt"
    }
  }
  ```

**Received Data:**

The data received in response to the client's sent message, is the `id` of the message.

```json
{
  "new_msg_id": 5,
}
```

The client is expected to perform an "optimistic UI rendering" of the message snippet with the data it uses to create the message, and set the message's `id` to the the one received in response, as soon as it is received. The full data needed for rendering on the partner's end is sent to it.

The `delivery status update` event described above, is how you recieve delivery status updates in order to update the current read receipt. It goes from `sent`, to `delivered`, and, finally, to `seen`.


### Open Group Messaging Stream

