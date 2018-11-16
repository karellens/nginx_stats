# Nginx stats parsers

```bash
go build per_day_counter.go
```

### Usage
```bash
./per_day_counter -get "ip uri" -source nginx-access.log -destination uri_views.json -from 2018-06-01 -to 2018-07-01 -pretty
```

You can combine the fields below with `+` to calculate the number of occurrences of each unique combination
```bash
./per_day_counter -get "ip+useragent uri" -source nginx-access.log -destination uri_views.json
```

Specify `=` before the combination if you want to display only the number of unique occurrences found
```bash
./per_day_counter -get "=ip+useragent uri" -source nginx-access.log -destination uri_views.json
```


### Available fields:
 - ip,
 - date,
 - datetime,
 - method,
 - uri,
 - query,
 - statuscode,
 - bytessent,
 - refferer,
 - useragent