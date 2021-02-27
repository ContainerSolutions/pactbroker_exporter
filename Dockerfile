FROM scratch
EXPOSE 9624
USER 1000

COPY pactbroker_exporter /bin/pactbroker_exporter

ENTRYPOINT ["pactbroker_exporter"]
