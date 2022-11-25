# GitHub codespaces devcontainer images

Images for use with GitHub codespaces and ACE v12. Built on top of the GitHub-provided
Ubuntu 22.04 standard image.

These images can be built and pushed to dockerhub or any other container registry, or
built as needed by the codespaces infrastructure. The pre-built container startup times
tend to be 20-60 seconds, while the container build can take five minutes or more due
to the ACE image download time.

## Dockerfiles

- [Dockerfile](Dockerfile) contains an ACE image that is English-only and does not include WSRR, but can be used for most ACE applications.
- [Dockerfile.mqclient](Dockerfile.mqclient) adds an MQ client on top of the main Dockerfile.
- [Dockerfile.ace-minimal](Dockerfile.ace-minimal) is a smaller image missing Adapters, global cache, and various other components. This image is used for the ACE demo pipeline but may need to be modified for other applications.

## Usage

Create a `devcontainer.json` file in a `.devcontainer` directory in the repo with
contents similar to the following:

```
{
    "name": "my-repo-devcontainer",
    "image": "my-container-registry/ace-minimal-devcontainer:12.0.4.0",
    "containerEnv": {
        "LICENSE": "accept"
    },
    "remoteEnv": {
        "REMOTE_LICENSE": "accept"
    }
}
```
replacing the image name with the location of the container built from one of the
Dockerfiles in this repo.

## Image sizes

The compressed size of the ace-minimal devcontainer is around 470MB, which is similar
to many of the GitHub-provided images at the time of writing: Python, Java 8, and
javascript-nodejs are in the 450-550MB range.

Full ACE installs (even with languages removed) are larger, with the main Dockerfile
image compressing to 670MB, and the MQ client increasing the size to 750MB. Despite
this, the observed image pull times from dockerhub are still less than a minue with
the US East region on github.com, though this is not in any way guaranteed.
