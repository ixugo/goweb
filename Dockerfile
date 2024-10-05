FROM gospace/alpine:latest

ENV TZ=Asia/Shanghai

RUN apk --no-cache add ca-certificates \
	tzdata

WORKDIR /src

ADD ./output/linux_amd64/app ./

LABEL Name=goweb Version=0.0.1

EXPOSE 8080

CMD [ "./app" ]