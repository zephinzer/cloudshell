# Cloudshell

This project contains an Xterm.js frontend that connets to a Go backend to provide a shell to the host system. Basically, access your shell from a browser.

Some use cases:

1. Deploy to a compute instance in your networks and expose it when needed to gain shell access to your network over the browser
2. Deploy to a Kubernetes cluster with appropriate (Cluster)Role and (Cluster)RoleBinding resources to allow some level of access to developers
3. Exposing a CLI tool (see [`./examples/k9s`](./examples/k9s) for an example) over the browser. Think CLI-as-a-frontend.

# Development

## Install dependencies

Run `make init` to install both Node.js and Golang dependencies.

## Test run it

Run `make start` to start the Go backend which will also serve the static files for the website.

Open your browser at http://localhost:8376 to view your shell in the browser.

# Build/Release

## Building the project

To build this project, run `make build` to build the binary

## Creating the Docker image

Run `make package` to create the Docker image. To customise the package process:

1. Create a file named `Makefile.properties` (this will be `-include`d by the Makefile)
2. Set `image_namespace` to your desired namespace (defaults to `zephinzer`)
3. Set `image_name` to your desired image name (defaults to `cloudshell`)
4. Set `image_tag` to your desired tag (defaults to the first 8 characters of the Git commit hash)

## Publishing the Docker image

Run `make publish` to publish the Docker image. Same customisations as above apply.

## Publishing example Docker images

Run `make publish-example id=${id}` to publish the example Docker images where `${id}` is the directory name of the directory in the [`./examples` directory](./examples).

# Deploy

## Running the Docker image

Run `make run` to run the Docker image locally

## Deploying via Helm

Go to [`./deploy/cloudshell`](./deploy/cloudshell) and run `helm install --values ./values-k9s.yaml --set-url url=cloudshell.yourdomainname.com cloudshell .`.

Modify the `values-k9s` file as required.

# License

This project is licensed under the MIT license.
