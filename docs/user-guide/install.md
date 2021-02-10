

## Install kubectl

=== "Linux"
    ```bash
    wget -nv -nc  https://dl.k8s.io/release/%%{ kubernetes.version }%%/bin/linx/amd64/kubectl && \
      chmod +x kubectl && \
      mv kubectl /usr/local/bin
    ```

=== "MacOSX"
    ```bash
    brew install kubectl
    ```

=== "Windows"
    ```bash
    wget -nv -nc -O https://dl.k8s.io/release/%%{ kubernetes.version }%%/bin/windows/amd64/kubectl
    ```


## Install stern

=== "Linux"
    ```bash
    wget -nv -nc -o stern   https://github.com/wercker/stern/releases/latest/download/stern_linux_amd64 && \
      chmod +x stern && \
      mv stern /usr/local/bin/
    ```

=== "MacOSX"
    ```zsh
    brew install stern
    ```

=== "Windows"
    ```zsh
    wget -nv -nc -o stern.exe   https://github.com/wercker/stern/releases/latest/download/stern_windows_amd64.exe
    ```


