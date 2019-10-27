FROM node:11 AS build-frontend

WORKDIR /tmp

ADD frontend/*.json ./
ADD frontend/*.lock ./
RUN yarn install --ignore-optionals

ADD frontend/. .
RUN yarn run lint
RUN yarn run build --prod

FROM golang AS build-server

WORKDIR /build

# copy dependency information and fetch them
COPY go.mod ./
RUN go mod download

# copy sources
COPY . .

# build and install (without C-support, otherwise there issue because of the musl glibc replacement on Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -a cmd/issues/issues.go

FROM alpine

# update CA certificates
RUN apk update && apk add ca-certificates postgresql-client
WORKDIR /usr/aybaze

COPY --from=build-frontend /tmp/dist ./frontend/dist
COPY --from=build-server /build/issues .

ADD restore.sh .
ADD docker-entrypoint.sh .
ADD sql sql

CMD ["./docker-entrypoint.sh"]
