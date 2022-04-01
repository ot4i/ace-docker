# updateBase Sample

## What is in the sample?

This sample applies updates all the base OS packages. This can be used to resolve CVEs in the base packages

## Building the sample

- First [build the ACE image](../README.md#Building-a-container-image) or obtain one of the shipped images
- In the `sample/scripts/updateBase` folder:

    ```bash
    docker build -t aceapp --build-arg FROMIMAGE=cp.icr.io/cp/appc/ace:12.0.4.0-r1 --file Dockerfile .
    ```
