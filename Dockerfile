# build apps
FROM alpine:latest AS lib-builder

RUN apk add g++ \
    pugixml-dev \
    openssl-dev \
    curl-dev \
    libzip-dev \
    make \
    bash \
    git

WORKDIR /usr/src

RUN git clone git://soutade.fr/libgourou.git \
  && cd libgourou \
  && make BUILD_STATIC=1

FROM golang:latest AS server-builder

WORKDIR /app

COPY server/go.mod ./
COPY server/main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o book-server

# copy from builder to runtime image
FROM alpine:latest

RUN apk add --no-cache \
  libcurl \
  libzip \
  pugixml \
  bash

COPY --from=lib-builder /usr/src/libgourou/utils/acsmdownloader \
                    /usr/src/libgourou/utils/adept_activate \
                    /usr/src/libgourou/utils/adept_remove \
                    /usr/local/bin/

COPY --from=server-builder /app/book-server /usr/local/bin/

WORKDIR /home/libgourou

#COPY scripts .

RUN mkdir -p output

EXPOSE 8080

#USER nonroot:nonroot

#ENTRYPOINT ["/bin/bash", "/home/libgourou/entrypoint.sh"]
ENTRYPOINT ["/usr/local/bin/book-server"]
