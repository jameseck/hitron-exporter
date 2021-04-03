FROM golang:1.16 as build

WORKDIR /src
COPY . .
RUN go build -ldflags="-w -s" -o /out/hitron-exporter .

FROM scratch
COPY --from=build /out/hitron-exporter /hitron-exporter

ENTRYPOINT [ "/hitron-exporter" ]
EXPOSE 80
