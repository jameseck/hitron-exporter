FROM golang:1.16 as build

ENV CGO_ENABLED=0
WORKDIR /src
COPY . .
RUN go build -ldflags="-w -s" -a -o /out/hitron-exporter .

FROM scratch
COPY --from=build /out/hitron-exporter /hitron-exporter

ENTRYPOINT [ "/hitron-exporter" ]
EXPOSE 80
