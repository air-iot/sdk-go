
go build -tags netgo -v -o go-flow-ext-demo main.go

docker build -t airiot/go-flow-ext-demo:v4.0.0 .

docker save airiot/go-flow-ext-demo:v4.0.0 | gzip > go-flow-ext-demo.tar.gz


