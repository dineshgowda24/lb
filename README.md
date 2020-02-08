# lb
application loadbalancer written in go

## Stating Application Servers

Go to `backendserver\config` and edit `config.json` to configure app servers
```bash
cd backendserver
go run server.go
```

## Starting Loadbalancer

In the root of the directory run the below commad
```bash
go run main.go
```
