# backend-subscribe

Client subscribe counterpart to backend-publish. Subscribe to receive the results of your requests to backend-publish in some weird extended streaming async HTTP-ish thing. Refer to ```backend-store```.

## API

```
GET /subscribe/:userid/client/:clientid
```

Subscribe to your SSE stream.

### URL Params

In the future, this will be supplied via token.

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| userid | String | User's ID. | ✓ |
| clientid | String | Device's ID. Must be unique to the device. I suggest something based on MAC address. | ✓ |

### Success Response (200 OK)

An [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource) stream.

### Event

```
{
  "code": <http status code>
  "message": <message>
}
```
