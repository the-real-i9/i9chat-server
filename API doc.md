# API Documentation

## API Error Codes

These are errors triggered by actions that are forbidden according to the business logic.
They often have reasons and explanations, which is why the below table is provided for you, the developer, to know which erroneous action you might be taking and how to handle it.

For example, trying to send a message to a user while specifying an invalid username that doesn't belong to any user in the application, or allowing an ordinary member of a group to take admin-restricted actions.

Although, it's rare that the final build of the frontend would allow this, however, a developer could make this mistake while implementing the API.

These error codes would appear in `400 Bad Request` error messages in this format `"business error: ERR_CODE"`.
