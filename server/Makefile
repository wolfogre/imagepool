build:
	go build

clear:
	rm -f server

docker: build
	docker build -t registry.cn-hangzhou.aliyuncs.com/wolfogre-hub/imagepool:2 .
	make clear


