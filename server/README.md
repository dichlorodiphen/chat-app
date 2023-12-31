# Server

## Websocket Endpoint

The websocket endpoint is located at `/ws`. After handshaking, the first message from the client should be a JWT (without the `Bearer ` prefix). After this token is verified by the server, the server will begin streaming messages to the client. If the token cannot be verifed, the server will close the websocket connection.

## REST API

### /users/signup (POST)

* Description: Creates a new account with the given credentials.
* Visibility: All
* Body:
    ```
    {
        username: <username>,
        password: <password>
    }     
    ```
* Responses:
    * 201 (CREATED)
        ```
        <JWT token>
        ```
    * 400 (BAD REQUEST)

### /users/login (POST)

* Description: Authenticates with given credentials.
* Visibility: All
* Body:
    ```
    {
        username: <username>,
        password: <password>
    } 
    ```
* Responses:
    * 200 (OK) - messing a bit with HTTP semantics but it's for the greater good
        ```
        <JWT token>
        ```
    * 400 (BAD REQUEST)
    * 403 (FORBIDDEN)

### /messages (GET)

* Description: Get all messages.
* Visibility: Authenticated
* Body: N/A
* Responses:
    * 200 (OK)
        ```
        {
            [
                {
                    id: <message id>,
                    author: <author username>,
                    content: <message content>,
                    votes: <votes>
                },
                ...
            ]
        } 
        ```
    * 401 (UNAUTHORIZED)

### /messages (POST)

* Description: Create a new message.
* Visibility: Authenticated
* Body:
    ```
    {
        content: <message content>
    }
    ```
* Responses:
    * 204 (NO CONTENT)
    * 401 (UNAUTHORIZED)
* Notes: Server should retrieve author username by extracting claims from JWT token.

### /messages/{id} (PATCH)

* Description: Update the vote count of an existing message.
* Visibility: Authenticated
* Body:
    ```
    {
        upvoted: <true or false>,
        downvoted: <true or false>
    } 
    ```
* Responses:
    * 204 (NO CONTENT)
    * 400 (BAD REQUEST)
    * 401 (UNAUTHORIZED)
    * 403 (FORBIDDEN)
    * 404 (NOT FOUND)
* Notes: Server should retrieve username by extracting claims from JWT token and handle vote logic to ensure there is no double-voting.
