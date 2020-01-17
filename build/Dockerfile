FROM ubuntu:xenial
RUN apt-get update
RUN apt-get install tzdata
RUN echo "Asia/Shanghai" > /etc/timezone
RUN rm -f /etc/localtime
RUN dpkg-reconfigure -f noninteractive tzdata
RUN apt-get install -y ca-certificates
#RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai/etc/localtime
#RUN ln -fs /usr/share/zoneinfo/US/Pacific-New /etc/localtime && dpkg-reconfigure -f noninteractive tzdata
WORKDIR /root/deploy/az-fin

RUN mkdir -p /root/deploy/az-fin
COPY config.toml /root/deploy/az-fin/config.toml
COPY golang-az-fin /root/deploy/az-fin/
RUN chmod +x /root/deploy/golang-az-fin/az-fin

ENTRYPOINT ["/root/deploy/az-fin/az-fin"]
