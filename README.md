# Nginx stats parsers

```bash
go build per_day_counter.go
```

Usage
```bash
./per_day_counter -get "ip uri" -source nginx-access.log -destination uri_views.json
```