# plurimos API

## Administrator APIs

### GET /mom/_api/apps

Get all available apps.

> Only `system` app can access this API.

### POST /mom/_api/app

Create a new app.

Input parameters:

```json
{
    "id": "(string, optional) app's unique id, if empty a random id will be generated",
    "secret": "(string) app's secret key, used for authentication",
    "any other arbitrary fields": "and arbitrary values"
}
```

Result: when successful, `result.status` is `200` and app's id is returned via `result.data`

```json
{
    "status": 200,
    "data": "<app-id>"
}
```

> Only `system` app can access this API.

### GET /mom/_api/app/:id

Get an app's info.

> Only "system" app and owner can request app info.

### PUT /mom/_api/app/:id

Update an existing app's info.

> Only "system" app and owner can update app info.

### DELETE /mom/_api/app/:id

Remove an existing app.

> Only `system` app can access this API.
