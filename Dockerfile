# Build do app que apaga imagens
FROM golang:1.19.0 as GO_COMPILE
WORKDIR /delete-images
COPY ./delete-images .
ENV GO111MODULE=off
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o delete-images .

#Monta o dpe-registry (docker registry + delete-images)
FROM registry:2.8.1
COPY README.md .
COPY --from=GO_COMPILE /delete-images/delete-images /delete-images/
