
# iFix Sample

## What is in the sample?

This sample applies an iFix provided by IBM to an existing image.

## Building the sample

- First [build the ACE image](../README.md#Building-a-container-image)  or obtain one of the shipped images
- Download the appropriate iFix provided by IBM and place it into the fix directory
- Update the dockerfile to reference the iFix i.e. replace \<iFixName> with the name of the ifix you've downloaded i.e. 12.0.3.0-ACE-LinuxX64-TFIT39515A
- In the `sample/scripts/applyIfix` folder:

    ```bash
    docker build -t aceapp --build-arg FROMIMAGE=cp.icr.io/cp/appc/ace:12.0.4.0-r1 --file Dockerfile .
    ```
