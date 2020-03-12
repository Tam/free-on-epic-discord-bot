# Free on Epic
A discord bot that tells you what the free games on the Epic Store are

```shell script
$ go get
$ go build main.go
$ ./main -url "[your channel webhook URL]"
```

Run as a cron (7pm every day)
```text
0 19 * * * ./main -url "[your channel webhook URL]" > /dev/null 2>&1
```
