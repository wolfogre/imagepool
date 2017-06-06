FROM centos:7

RUN rm /etc/localtime && ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY server /opt/server

CMD ["/opt/server"]
