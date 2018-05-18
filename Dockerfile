FROM golang AS builder

COPY . $GOPATH
WORKDIR $GOPATH
RUN set -ex \
  && go build -o /go/agent src/main/main.go

FROM registry.cn-hangzhou.aliyuncs.com/aliware2018/services AS exist

FROM registry.cn-hangzhou.aliyuncs.com/aliware2018/debian-jdk8

COPY --from=exist /root/workspace/services/mesh-provider/target/mesh-provider-1.0-SNAPSHOT.jar /root/dists/mesh-provider.jar
COPY --from=exist /root/workspace/services/mesh-consumer/target/mesh-consumer-1.0-SNAPSHOT.jar /root/dists/mesh-consumer.jar
COPY --from=builder /go/agent /root/dists/agent
COPY --from=exist /usr/local/bin/docker-entrypoint.sh /usr/local/bin
COPY start-agent.sh /usr/local/bin

RUN set -ex \
 && chmod a+x /usr/local/bin/start-agent.sh \
 && mkdir -p /root/logs

EXPOSE 8087

ENTRYPOINT ["docker-entrypoint.sh"]
