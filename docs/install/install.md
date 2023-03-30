# Install

## Download the latest binary release :material-github:
---

=== "Darwin :material-apple:"
    ```
    curl -sL https://github.com/gmeghnag/omc/releases/latest/download/omc_Darwin_x86_64.tar.gz | tar xzf - omc
    chmod +x ./omc
    ```
=== "Linux :simple-linux:"
    ``` aml
    curl -sL https://github.com/gmeghnag/omc/releases/latest/download/omc_Linux_x86_64.tar.gz | tar xzf - omc
    chmod +x ./omc   
    ```
=== "Windows :fontawesome-brands-windows:"
    ``` 
    curl.exe -sL "https://github.com/gmeghnag/omc/releases/latest/download/omc_Windows_x86_64.zip" -o omc.zip 
    tar -xf omc.zip
    ./omc.exe 
    ```


## Build from the source code :fontawesome-brands-golang:
---
```
git clone https://github.com/gmeghnag/omc.git
cd omc/
go install
```
