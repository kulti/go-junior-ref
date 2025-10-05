FROM migrate/migrate:v4.19.0

RUN apk add postgresql-client

COPY bin /app/bin
COPY migrations /app/migrations

ENTRYPOINT [ "/app/bin/entrypoint" ]
CMD [ "migrate", "up" ]
