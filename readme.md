# Buildpack Pack Builder

This project is designed to allow for an expected buildpack build environment, even when you're on a different CPU architecture or OS than the target system. This also allows for building multiple buildpacks for online or offline (cached) use asynchronously to save time. Some of these buildpacks take a decent amount of time! I have been able to use this to build all public buildpacks in offline mode within about 30 minutes total.

## Requirements

- Docker on any OS

## Building the Docker for first time use

From the root of this repository (parent to the src directory) you will need to build the Docker image. The Docker image can be build using the following command:

` docker build -t buildpackpackbuilder:latest . `

## How To Use

This Docker image is meant to be run as an application in itself, not as a server or as a service.

The container requires two parts to be supplied through volumes:

- A build directory (this is where the build output goes). This will be **/build**
- A Yaml file that details which buildpacks to build and how to build them. This needs to be at the root and can be any name. The name of the yaml file also need to be passed as a parameter.

The yaml file itself is expected to be at the root of the container, but can be any name. The container will expect a parameter of the yaml file to be used.

Here is an example command running the container and voluming in the build and yaml configuration.

` docker run -it -v "/path/to/local/build/directory/:/build/" -v "/path/to/yaml/example.yml:example.yml" buildpackpackbuilder:latest example.yml `

Once it begins running, it will read out each pack that is being built asynchronously. Once building is finished, the containter will automatically stop.

## Configuration

*See **example.yml** in this repository for examples. This are all currently building as is.*

The yaml should contain an array of desired buildpacks.

```yaml
buildpacks:
  - name: nodejs-buildpack
    version: 1.7.70
    stack: cflinuxfs3
    official: yes
    type: tar
    offline: yes
    skip: no
    build:
      type: packager
```

Each buildpack will require the above at the very least. This is the most basic buildpack configuration.

- **name** - This is the name of the buildpack per the official repository
- **version** - This is the required version. This must be an existing version on the official repository for the buildpack.
- **stack** - This is the requested stack. This can be either cflinuxfs3 or windows or whatever the buildpack allows.
- **official** - This helps the builder determin if this will come from the official github or if another URI or location will be provided
- **type** - This defines the expected file type. Currently this will always be **tar**
- **tar** - If a different location than the official github repo is requested, the url can be supplied under tar - url.
- **skip** - If set to yes, this allows for this buildpack to be defined but skipped or ignored when run.
- **build** - This section helps define different ways to build the buildpack. 
- **build - type** - A large majority will use **packager** with a few exceptions. Java will need to be defined as **java**
