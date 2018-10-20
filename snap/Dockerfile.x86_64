FROM snapcore/snapcraft:stable
ADD . /code
WORKDIR /code
ENV SNAPCRAFT_SETUP_CORE=1
ENV KUBE_ARCH=$KUBE_ARCH
ENV KUBE_VERSION=$KUBE_VERSION
RUN apt-get update
RUN apt-get -y install wget curl file make
ENTRYPOINT ["./build-scripts/build"]
CMD ["$@"]
