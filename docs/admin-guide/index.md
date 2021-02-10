
# Install karina

=== "Linux"
    ```bash
    wget -nv -nc -O karina \
      https://github.com/flanksource/karina/releases/latest/download/karina && \
      chmod +x karina && \
      mv karina /usr/local/bin/karina
    ```

=== "MacOSX"
    ```zsh
    wget -nv -nc -O karina \
      https://github.com/flanksource/karina/releases/latest/download/karina_osx && \
      chmod +x karina && \
      mv karina /usr/local/bin/karina
    ```


!!! note
    For production pipelines you should always pin the version of karina you are using
