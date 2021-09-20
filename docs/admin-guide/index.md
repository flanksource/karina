
# Install karina

=== "Linux"
    ```bash
    wget -nv -O karina \
      https://github.com/flanksource/karina/releases/latest/download/karina_linux-amd64 && \
      chmod +x karina && \
      mv karina /usr/local/bin/karina
    ```

=== "MacOSX"
    ```zsh
    wget -nv -O karina \
      https://github.com/flanksource/karina/releases/latest/download/karina_darwin-amd64 && \
      chmod +x karina && \
      mv karina /usr/local/bin/karina
    ```


!!! note
    For production pipelines you should always pin the version of karina you are using
