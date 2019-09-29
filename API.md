# plurimos API

## Administrator APIs

### GET /mom/_api/apps

Get all available apps.

Output

```json
{
    "status": 200,
    "data": [
        { app-data-1 },
        { app-data-2 },
        ...
    ]
}
```

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

Output: when successful, `status` is `200` and app's id is returned via `data`.

```json
{
    "status": 200,
    "data": "<app-id>"
}
```

> Only `system` app can access this API.

### GET /mom/_api/app/:id

Get an app's info.

Input parameters:

- `id`: app's unique id, passed to API via url path.

Output:

```json
{
    "status": 200,
    "data": { app-data }
}
```

> Only "system" app and owner can request app info.

### PUT /mom/_api/app/:id

Update an existing app's info.

Input parameters:

- `id`: app's unique id, passed to API via url path.
- App's data, passed via request body as a JSON.

```json
{
    "secret": "(optional, string) app's new secret key",
    "any other arbitrary fields": "and arbitrary values"
}
```

Output: when successful, `status` is `200`.

```json
{
    "status": 200
}
```

> Only "system" app and owner can update app info.

### DELETE /mom/_api/app/:id

Remove an existing app.

Input parameters:

- `id`: app's unique id, passed to API via url path.

Output: when successful, `status` is `200`.

```json
{
    "status": 200
}
```

> Only `system` app can access this API.

## Mapping APIs

### GET /mom/api/:ns/:from

Get an existing mapping

Input parameters:

- `ns`: namespace, passed to API via url path.
- `from`: object to check if it has been mapped to any target within the namespace, passed to API via url path.

Output: if mapping not found `status` is `404`; if mapping found `status` is `200` and mapping info is returned via `data`.

```json
{
    "status": 200,
    "data": {
        "ns" : "namespace",
        "frm": "object",
        "to" : "target",
        "t"  : "timestamp, example 2019-09-28T16:17:37+07:00",
        "app": "app-id (optional)"
    }
}
```

### PUT /mom/api/:ns/:from/:to

Map an object (:from) to a target (:to).

Input parameters:

- `ns`: namespace, passed to API via url path.
- `from`: the object to map, passed to API via url path.
- `to`: the target to map, passed to API via url path.

Output: when successful, `status` is `200` and mapping info is returned via `data`.

```json
{
    "status": 200,
    "data": {
        "ns" : "namespace",
        "frm": "object",
        "to" : "target",
        "t"  : "timestamp, example 2019-09-28T16:17:37+07:00",
        "app": "app-id (optional)"
    }
}
```

### DELETE /mom/api/:ns/:from/:to

Unmap an existing mapping.

Input parameters:

- `ns`: namespace, passed to API via url path.
- `from`: the object to unmap, passed to API via url path.
- `to`: the target to unmap, passed to API via url path.

Output: when successful, `status` is `200`

```json
{
    "status": 200,
    "message": "Ok"
}
```

### GET /mom/api/_/:to?ns=<namespace-list>

Get reversed mappings of a target (:to).

Input parameters:

- `to`: the target, passed to API via url path.
- `ns`: namespace list, separated by comma (,) or semi-colon (;), passed to API via url query.

Output: when successful, `status` is `200` and mapping data is returned via `data`.

```json
{
    "status": 200
    "data": {
        "namespace-1": [
            { mapping-data-1 },
            { mapping-data-2 },
            ...
        ],
        "namespace-2": [
            { mapping-data-1 },
            { mapping-data-2 },
            ...
        ],
        ...
}
```

### POST /mom/api/_

Performs bulk mapping from objects to a target on multiple namespaces.

Input parameters: a map of `{namespace:object}` in request body.

Business rules:

- All specified objects will map to a same target.
- If non of specified objects is currently mapping to any target, a random target is generated for mapping.
- If some of specified objects are currently mapping to a target, this target is used to map to other objects.
- If there are two of the specified objects are currently mapping to different targets, API fails with status `409-conflict`.

Output: when successful, `status` is `200` and target is returned via `data`.

```json
{
    "status": 200,
    "data": "target"
}
```
