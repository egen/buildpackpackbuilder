# DOCS TBD

## Build
From the root of the source directory:
` docker build -t buildpackpackbuilder:latest . `

## How To Use

` docker run -it -v "/path/to/local/build/directory/:/build/" buildpackpackbuilder:dev example.yml `