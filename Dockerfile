FROM busybox:latest
COPY .bin/platform-cli /bin/
ENTRYPOINT [ "/bin/platform-cli" ]
