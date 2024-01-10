#
# file Dockerfile
#
# Base docker image with all required dependencies for OB-USP-A
#
# Based on Ubuntu 22.10 (Kinetic Kudu), which provides libmosquitto 2.0.11 and libwebsockets 4.1.6
# This image includes some basic compilation tools (automake, autoconf)
#
# One-liner execution line (straightforward build for OB-USP-A execution):
# > docker build -f Dockerfile -t obuspa:latest .
#
# Multi-stage builds execution lines (to tag build stages):
# 1) Create the build environment image:
# > docker build -f Dockerfile -t obuspa:build-env --target build-env .
# 2) Create the OB-USP-A image, then build the application
# > docker build -f Dockerfile -t obuspa:latest --target runner .
#
#FROM --platform=$BUILDPLATFORM ubuntu AS build-env
FROM ubuntu AS build-env

# Install dependencies
RUN apt update && apt -y install \
        build-essential \
        libssl-dev \
        libcurl4-openssl-dev\
        libsqlite3-dev \
        libz-dev \
        autoconf \
        automake \
        libtool \
        libmosquitto-dev \
        pkg-config \
        git \
        cmake \
        make \
    && apt clean

RUN mkdir -p /usr/local/src
WORKDIR /usr/local/src/
RUN git clone https://github.com/warmcat/libwebsockets.git libwebsockets
WORKDIR /usr/local/src/libwebsockets
RUN cmake -B build -S .
RUN cd build && make && make install
# install libs in /usr/local/lib ; configured in /etc/ld.so.conf.d/libc.conf
# ENV LD_LIBRARY_PATH /usr/local/lib:${LD_LIBRARY_PATH}
RUN ldconfig -v

ENV MAKE_JOBS=8

WORKDIR /

RUN git clone https://github.com/BroadbandForum/obuspa.git

RUN cd /obuspa/ && \
    autoreconf -fi && \
    ./configure && \
    make -j${MAKE_JOBS} && \
    make install

FROM debian:stable AS build-release-stage
RUN apt update && apt -y install \
    libssl-dev \
    libsqlite3-dev \
    libcurl4-openssl-dev\
    libmosquitto-dev

WORKDIR /
COPY --from=build-env /usr/local/lib/libwebsockets.* /usr/local/lib/
COPY --from=build-env /obuspa/obuspa /bin
RUN mkdir usr/local/var/
RUN mkdir usr/local/var/obuspa
RUN ldconfig -v