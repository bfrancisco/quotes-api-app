# Quotes API Test Commands

Use these commands on cmd. 
(Only tested on Windows)

## Start the API

```cmd
go run ./rest-api
```

Base URL:

```cmd
set BASE_URL=http://localhost:8080/v1
```

The running server starts with an empty in-memory store. Create a quote before testing data-dependent read, update, and delete requests.

## Happy Path

Health check:

```cmd
curl.exe "%BASE_URL%/health"
```

List quotes from an empty store:

```cmd
curl.exe "%BASE_URL%/quotes"
```

List quotes with pagination:

```cmd
curl.exe "%BASE_URL%/quotes?limit=10&offset=0"
```

Filter quotes by author:

```cmd
curl.exe "%BASE_URL%/quotes?author=Edsger+W.+Dijkstra"
```

Create a quote:

```cmd
curl.exe -X POST "%BASE_URL%/quotes" ^
  -H "Content-Type: application/json" ^
  -d "{\"text\":\"Simplicity is prerequisite for reliability.\",\"author\":\"Edsger W. Dijkstra\"}"
```

Get quote by ID:

```cmd
curl.exe "%BASE_URL%/quotes/1"
```

Get a random quote:

```cmd
curl.exe "%BASE_URL%/quotes/random"
```

Update only the text:

```cmd
curl.exe -X PATCH "%BASE_URL%/quotes/1" ^
  -H "Content-Type: application/json" ^
  -d "{\"text\":\"Simplicity is a prerequisite for reliability.\"}"
```

Update only the author:

```cmd
curl.exe -X PATCH "%BASE_URL%/quotes/1" ^
  -H "Content-Type: application/json" ^
  -d "{\"author\":\"Edsger Dijkstra\"}"
```

Delete a quote:

```cmd
curl.exe -i -X DELETE "%BASE_URL%/quotes/1"
```

## Error Checks

Get a quote with an invalid ID:

```cmd
curl.exe "%BASE_URL%/quotes/not-a-number"
```

Create a quote with invalid text:

```cmd
curl.exe -X POST "%BASE_URL%/quotes" ^
  -H "Content-Type: application/json" ^
  -d "{\"text\":\"\",\"author\":\"Someone\"}"
```

Patch with no fields to update:

```cmd
curl.exe -X PATCH "%BASE_URL%/quotes/1" ^
  -H "Content-Type: application/json" ^
  -d "{}"
```

Request a random quote before creating one:

```cmd
curl.exe "%BASE_URL%/quotes/random"
```

## Suggested Order

Run the commands in this order on a fresh server:

1. `curl.exe "%BASE_URL%/health"`
2. `curl.exe "%BASE_URL%/quotes"`
3. Create a quote with `POST /quotes`
4. `GET /quotes/1`
5. `GET /quotes/random`
6. `PATCH /quotes/1`
7. `DELETE /quotes/1`
