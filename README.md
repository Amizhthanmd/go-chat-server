# Go Chat Server API Documentation

## **Getting Started**

Clone the repository and set up the project.

```bash
git clone https://github.com/Amizhthanmd/go-chat-server.git

cd go-chat-server

go mod tidy

go run main.go
```

## **User Management**

### Add User

Create a new user with a display name.

**Endpoint**: `POST /api/v1/users`

**Request Body**:

```json
{
  "display_name": "abc"
}
```

### Get All Users

Retrieve a list of all registered users.

**Endpoint**: `GET /api/v1/users`

---

## **Direct Messaging**

### Open Chat with User

Open a chat with a specific user by their unique ID.

**Endpoint**: `GET /api/v1/users/chat`

**Query Parameters**:

- `user_id`: Unique ID of the user to chat with (e.g., `123-456-789`).

### Send Message to User

Send a message directly to another user.

**Endpoint**: `POST /api/v1/users/chat`

**Request Body**:

```json
{
  "sender_id": "e50299fc-c365-47fe-97bc-30e593b042c8",
  "receiver_id": "390a428d-6fd9-46b7-b46f-e7a9ca53922d",
  "message": "hai"
}
```

---

## **Room Management**

### Create Room

Create a new chat room with a specified name.

**Endpoint**: `POST /api/v1/rooms`

**Request Body**:

```json
{
  "room_name": "xyz"
}
```

### Join Room

Join a chat room by providing the room ID and user ID.

**Endpoint**: `GET /api/v1/rooms/join`

**Request Body**:

```json
{
  "room_id": "Qsd6-uuFb-97Fl",
  "user_id": "30a19044-99b7-4b4a-8ce5-458fa3763396"
}
```

### Get Users in Room

Retrieve a list of all users currently in a specified room.

**Endpoint**: `GET /api/v1/rooms/users`

**Query Parameters**:

- `room_id`: Unique ID of the room (e.g., `Qsd6-uuFb-97Fl`).
- `user_id`: Unique ID of the requesting user (e.g., `36ee7f66-d404-445e-9e49-7011b5412d0a`).

**Example**:

```
GET /api/v1/rooms/users?room_id=Qsd6-uuFb-97Fl&user_id=36ee7f66-d404-445e-9e49-7011b5412d0a
```

### Send Message to Room

Send a message to all users in a specified chat room.

**Endpoint**: `POST /api/v1/rooms/send`

**Query Parameters**:

- `room_id`: Unique ID of the room (e.g., `Qsd6-uuFb-97Fl`).

**Request Body**:

```json
{
  "user_id": "36ee7f66-d404-445e-9e49-7011b5412d0a",
  "message": "hello welcome"
}
```
