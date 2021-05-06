ARG IMAGE_NAMESPACE=zephinzer
ARG IMAGE_NAME=cloudshell
ARG IMAGE_TAG=latest
FROM ${IMAGE_NAMESPACE}/${IMAGE_NAME}:${IMAGE_TAG}
ARG K9S_VERSION=v0.24.9
USER root
RUN curl -fL "https://github.com/derailed/k9s/releases/download/${K9S_VERSION}/k9s_Linux_x86_64.tar.gz" \
		  | tar xfz - -C "/tmp/" \
    && mv /tmp/k9s /usr/bin/k9s
USER user
CMD ["-t", "k9s", "-r", "--readonly"]
