# Issues [![Actions Status](https://github.com/aybaze/issues/workflows/build/badge.svg)](https://github.com/aybaze/issues/actions) [![](https://godoc.org/github.com/aybaze/issues?status.svg)](http://godoc.org/github.com/aybaze/issues)

`Issues` is a simple GitHub App that we use at Aybaze to organize issues across multiple repositories.

## Development

Start `postgres` and import `sql/issues.sql`.

```
docker run -e POSTGRES_DB=issues -d -p 5432:5432 postgres
psql -d issues -h localhost -U postgres < sql/issues.sql
```

Start `ngrok http 8000` to tunnel port 8000 to a public endpoint for the development application.