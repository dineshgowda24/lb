# lb
application loadbalancer written in go

## Stating Application Servers

Go to `backendserver\config` and edit `config.json` to configure app servers.
A sample configuration would look like:
`{
  "Servers" : [
    {"Host" : "localhost", "Port" : "8080", "Weight" : 80},
    {"Host" : "localhost", "Port" : "8088", "Weight" : 80},
    {"Host" : "localhost", "Port" : "8082", "Weight" : 80},
    {"Host" : "localhost", "Port" : "8083", "Weight" : 80},
    {"Host" : "localhost", "Port" : "8084", "Weight" : 80},
    {"Host" : "localhost", "Port" : "8085", "Weight" : 80},
    {"Host" : "localhost", "Port" : "8086", "Weight" : 80},
    {"Host" : "localhost", "Port" : "8087", "Weight" : 80}
  ]
}`
```bash
cd backendserver
go run server.go
```

## Starting Loadbalancer

In the root of the directory run the below commad
```bash
go run main.go
```
