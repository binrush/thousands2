# Thousands2 API Documentation

## Base URL
The API is served at the root path of the application.

## Authentication
Some endpoints require authentication. When authentication is required, the API will return a 401 Unauthorized status code with the message "Authentication required".

The API supports OAuth 2.0 authentication through VK (VKontakte) social network.

## Authentication Endpoints

### 1. OAuth Login

#### GET /auth/oauth/{provider}
Initiates the OAuth login flow for the specified provider.

**Path Parameters:**
- `provider`: string (currently only "vk" is supported)

**Response:**
- 302 Redirect to the OAuth provider's login page
- 404 Not Found if provider is not supported
- 302 Redirect to /user/me if user is already logged in

### 2. OAuth Callback

#### GET /auth/authorized/{provider}
Handles the OAuth callback after successful authentication.

**Path Parameters:**
- `provider`: string (currently only "vk" is supported)

**Query Parameters:**
- `code`: string (OAuth authorization code)
- `state`: string (OAuth state parameter for CSRF protection)

**Response:**
- 302 Redirect to /user/me on successful authentication
- 400 Bad Request if OAuth flow fails
- 500 Internal Server Error on server errors

### 3. Logout

#### GET /auth/logout
Logs out the current user and destroys their session.

**Response:**
- 302 Redirect to the root path (/)
- 500 Internal Server Error if session destruction fails

## Endpoints

### 1. Summit Endpoints

#### GET /summit/{ridgeId}/{summitId}
Retrieves detailed information about a specific summit.

**Response:**
```json
{
  "id": "string",
  "name": "string | null",
  "name_alt": "string | null",
  "interpretation": "string | null",
  "description": "string | null",
  "height": "integer",
  "coordinates": [float32, float32],
  "ridge": {
    "id": "string",
    "name": "string",
    "color": "string"
  },
  "images": [
    {
      "url": "string",
      "comment": "string"
    }
  ]
}
```

#### PUT /summit/{ridgeId}/{summitId}
Updates a user's climb record for a specific summit. Requires authentication.

**Request Body:**
- `comment`: string (optional)
- `date`: string (format: "DD.MM.YYYY", "MM.YYYY", or "YYYY")

**Response:**
- 200 OK on success
- 401 Unauthorized if not authenticated
- 400 Bad Request if date format is invalid
- 500 Internal Server Error on server errors

### 2. Summits Endpoint

#### GET /summits
Retrieves a list of all summits with additional information about user's climbs.

**Response:**
```json
{
  "summits": [
    {
      "id": "string",
      "name": "string | null",
      "height": "integer",
      "lat": "float32",
      "lng": "float32",
      "ridge": "string",
      "ridge_id": "string",
      "color": "string",
      "visitors": "integer",
      "rank": "integer",
      "is_main": "boolean",
      "climbed": "boolean"
    }
  ]
}
```

### 3. Top Climbers Endpoint

#### GET /top
Retrieves a paginated list of top climbers.

**Query Parameters:**
- `page`: integer (optional, defaults to 1)

**Response:**
```json
{
  "items": [
    {
      "user_id": "integer",
      "user_name": "string",
      "climbs_num": "integer",
      "place": "integer"
    }
  ],
  "page": "integer",
  "total_pages": "integer"
}
```

### 4. User Endpoints

#### GET /user/{userId}
Retrieves information about a specific user.

**Path Parameters:**
- `userId`: integer or "me" (to get current user's information)

**Response:**
```json
{
  "id": "integer",
  "oauth_id": "string",
  "src": "integer",
  "name": "string"
}
```

## Error Responses

The API uses consistent error responses with the following format:

```json
{
  "error": "string"
}
```

Common error status codes:
- 400 Bad Request: Invalid request parameters
- 401 Unauthorized: Authentication required
- 404 Not Found: Resource not found
- 405 Method Not Allowed: HTTP method not supported
- 500 Internal Server Error: Server-side error

## Data Types

### InexactDate
A flexible date format that can represent:
- Full date (DD.MM.YYYY)
- Month and year (MM.YYYY)
- Year only (YYYY)

### Coordinates
Array of two float32 values representing [latitude, longitude]

### SummitImage
```json
{
  "url": "string",
  "comment": "string"
}
```

### Ridge
```json
{
  "id": "string",
  "name": "string",
  "color": "string"
}
```

This API provides a comprehensive interface for managing mountain summit data, user climbs, and rankings. It supports both authenticated and unauthenticated access, with appropriate error handling and consistent response formats. 