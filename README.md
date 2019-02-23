# backend-subscribe

Client subscribe counterpart to backend-publish. Subscribe to receive the results of your requests to backend-publish in some weird extended streaming async HTTP-ish thing. Refer to ```backend-store```.

**To run this service securely means to run it behind traefik forwarding auth to `backend-auth`**

## Environment Variables

Supply environment variables by either exporting them or editing ```.env```.

| ENV | Description | Default |
| ---- | ----------- | ------- |
| LISTEN | Host and port number to listen on | :8080 |
| NATS | Host and port of nats | nats://localhost:4222 |

## API

```
GET /subscribe
```

Subscribe to your SSE stream.

### Required headers

| Name | Description |
| ---- | ----------- |
| X-User-Claim | Stringified user claim, populated by `backend-auth` called by `traefik` |

### Success Response (200 OK)

An [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource) stream.

### Event

```
{
  "code": <http status code>
  "message": <message>
}
```

#### Errors

| Code | Description |
| ---- | ----------- |
| 400 | Invalid user claims header. |
