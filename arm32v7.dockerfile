FROM golang:1.12 AS build
ENV GOARCH arm
ENV GOARM 7
ENV CGO_ENABLED 0
ENV GOOS linux
RUN mkdir /app
RUN mkdir /build
ADD . /build
WORKDIR /build
RUN go mod download
RUN go build -ldflags="-w -s" -a -installsuffix cgo -o /app/proxy main.go

FROM alpine AS helper
ENV GROUP appgroup
ENV USER appuser
COPY --from=build /app /app
RUN addgroup ${GROUP} && adduser -D ${USER} -G ${GROUP}
RUN chmod +x /app/proxy
RUN chown -R ${USER}:${GROUP} /app/

FROM scratch
COPY --from=helper /etc/passwd /etc/passwd
COPY --from=helper /etc/group /etc/group
COPY --from=helper /app /app

USER ${USER}
WORKDIR /app
CMD ["./proxy"]