# Holler
Holler is a HTTP/1/1.1/2 reverse proxy which supports dynamic backend registration

## Run 
```
go run cmd/hollar/main.go
```

make a dumb backend
```
cd ext && python -m SimpleHTTPServer 9001
```

register new backend
```
curl -XPOST -d @post.json localhost:9000/register/backend
```

make a request to the backend via proxy/foo
```
curl localhost:9000/foo
```

## TODO
- Autogenerate swagger-like spec from `description` fields of the dynamically registered service
- HTTP/1/1.1 (MVP)
- HTTP/2 (follow on work)
- gRPC Proxy (stretch goal)

