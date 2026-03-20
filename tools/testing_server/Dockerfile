FROM alpine:3.20

RUN apk add --no-cache netcat-openbsd bash

COPY ./tools/testing_server/testing_server.sh /test_client.sh
RUN chmod +x /test_client.sh

ENTRYPOINT ["/bin/sh"]
