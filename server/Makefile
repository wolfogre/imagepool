build:
	go build

clear:
	rm server

docker: build
	docker build -t registry.cn-hangzhou.aliyuncs.com/wolfogre-hub/imagepool:1 . 
	make clear


