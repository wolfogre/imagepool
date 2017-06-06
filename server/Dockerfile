FROM centos:7

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 && \
     rm /etc/localtime && ln -s /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime

COPY server /opt/server

CMD ["/opt/server"]
