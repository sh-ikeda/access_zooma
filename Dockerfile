FROM golang:1.13.7-alpine3.11 AS build-env

WORKDIR /go/build

COPY . .
WORKDIR /go/build/cmd/gen_zooma_query/
RUN go build
WORKDIR /go/build/cmd/get_bs_json/
RUN go build
WORKDIR /go/build/cmd/query_zooma/
RUN go build


FROM alpine:3.11

COPY --from=build-env /go/build/cmd/gen_zooma_query/gen_zooma_query /usr/local/bin
COPY --from=build-env /go/build/cmd/get_bs_json/get_bs_json /usr/local/bin
COPY --from=build-env /go/build/cmd/query_zooma/query_zooma /usr/local/bin

CMD ["/bin/sh"]
