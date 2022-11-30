FROM ubuntu:latest

FROM neo4j:4.4.15-community

### base ###
RUN yes | unminimize \
    && apt-get install -yq \
        asciidoctor \
        bash-completion \
        build-essential \
        htop \
        jq \
        less \
        llvm \
        locales \
        man-db \
        nano \
        software-properties-common \
        sudo \
        vim \
    && locale-gen en_US.UTF-8 \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/*
ENV LANG=en_US.UTF-8

### Gitpod user ###
# '-l': see https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#user
RUN useradd -l -u 33333 -G sudo -md /home/gitpod -s /bin/bash -p gitpod gitpod \
    # passwordless sudo for users in the 'sudo' group
    && sed -i.bkp -e 's/%sudo\s\+ALL=(ALL\(:ALL\)\?)\s\+ALL/%sudo ALL=NOPASSWD:ALL/g' /etc/sudoers
ENV HOME=/home/gitpod
WORKDIR $HOME
# custom Bash prompt
RUN { echo && echo "PS1='\[\e]0;\u \w\a\]\[\033[01;32m\]\u\[\033[00m\] \[\033[01;34m\]\w\[\033[00m\] \\\$ '" ; } >> .bashrc


### Yarn ###
RUN curl -fsSL https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - \
    && apt-add-repository -yu "deb https://dl.yarnpkg.com/debian/ stable main" \
    && apt-get install --no-install-recommends -yq yarn=1.12.3-1 \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/*

### Gitpod user (2) ###
USER gitpod
# use sudo so that user does not get sudo usage info on (the first) login
RUN sudo echo "Running 'sudo' for Gitpod: success"

### Go ###
ENV GO_VERSION=1.19.3 \
    GOPATH=$HOME/go-packages \
    GOROOT=$HOME/go
ENV PATH=$GOROOT/bin:$GOPATH/bin:$PATH
RUN curl -fsSL https://storage.googleapis.com/golang/go$GO_VERSION.linux-amd64.tar.gz | tar -xzv \
    && go get -u -v \
        github.com/acroca/go-symbols \
        github.com/cweill/gotests/... \
        github.com/davidrjenni/reftools/cmd/fillstruct \
        github.com/fatih/gomodifytags \
        github.com/haya14busa/goplay/cmd/goplay \
        github.com/josharian/impl \
        github.com/nsf/gocode \
        github.com/ramya-rao-a/go-outline \
        github.com/rogpeppe/godef \
        github.com/uudashr/gopkgs/cmd/gopkgs \
        github.com/zmb3/gogetdoc \
        golang.org/x/lint/golint \
        golang.org/x/tools/cmd/godoc \
        golang.org/x/tools/cmd/gorename \
        golang.org/x/tools/cmd/guru \
        sourcegraph.com/sqs/goreturns
# user Go packages
ENV GOPATH=/workspace:$GOPATH \
    PATH=/workspace/bin:$PATH


### Install Java
RUN bash -c ". /home/gitpod/.sdkman/bin/sdkman-init.sh && \
    sdk install java 18.0.2-open && \
    sdk default java 18.0.2-open"


### checks ###
# no root-owned files in the home directory
RUN notOwnedFile=$(find . -not "(" -user gitpod -and -group gitpod ")" -print -quit) \
    && { [ -z "$notOwnedFile" ] \
        || { echo "Error: not all files/dirs in $HOME are owned by 'gitpod' user & group"; exit 1; } }


## Install Neo4J latest version
RUN wget -O - https://debian.neo4j.com/neotechnology.gpg.key | sudo apt-key add - && \
	echo 'deb https://debian.neo4j.com stable latest' | sudo tee /etc/apt/sources.list.d/neo4j.list && \
	apt-get update && \
	apt-get install neo4j



## set neo4j environment variable
ENV NEO4J_AUTH=neo4j/test
