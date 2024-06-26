FROM golang:1.21.6-bullseye

SHELL ["/bin/bash", "-c"]

RUN apt-get update && apt-get -y upgrade \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get -y clean

WORKDIR /build

COPY . .
RUN go mod tidy \
    && go mod download \
    && go build -o /extproc


FROM busybox

COPY --from=0 /extproc /bin/extproc
RUN chmod +x /bin/extproc

ARG EXAMPLE=anti-replay

EXPOSE 50051

ENTRYPOINT [ "/bin/extproc" ]
CMD [ "anti-replay", "--log-stream", "--log-phases", "timespan", "60"  ]
