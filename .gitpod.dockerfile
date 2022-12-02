FROM ubuntu:latest

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
        curl \
        git \
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
ENV GOPROXY=direct
RUN curl -fsSL https://storage.googleapis.com/golang/go$GO_VERSION.linux-amd64.tar.gz | tar -xzv \
    && echo "Go version: $(go version)" && go install -v github.com/acroca/go-symbols@latest \
            && go install -v github.com/cweill/gotests/gotests@latest \
            && go install -v github.com/davidrjenni/reftools/cmd/fillstruct@latest \
            && go install -v github.com/fatih/gomodifytags@latest \
            && go install -v github.com/haya14busa/goplay/cmd/goplay@latest \
            && go install -v github.com/josharian/impl@latest \
            && go install -v github.com/nsf/gocode@latest \
            && go install -v github.com/ramya-rao-a/go-outline@latest \
            && go install -v github.com/rogpeppe/godef@latest \
            && go install -v github.com/zmb3/gogetdoc@latest \
            && go install -v github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest \
            && go install -v golang.org/x/lint/golint@latest \
            && go install -v golang.org/x/tools/cmd/godoc@latest \
            && go install -v golang.org/x/tools/cmd/gorename@latest \
            && go install -v golang.org/x/tools/cmd/guru@latest \
            && go install -v sourcegraph.com/sqs/goreturns@latest
# user Go packages
ENV GOPATH=/workspace:$GOPATH \
    PATH=/workspace/bin:$PATH

### checks ###
# no root-owned files in the home directory
RUN notOwnedFile=$(find . -not "(" -user gitpod -and -group gitpod ")" -print -quit) \
    && { [ -z "$notOwnedFile" ] \
        || { echo "Error: not all files/dirs in $HOME are owned by 'gitpod' user & group"; exit 1; } }

## set neo4j environment variable
ENV NEO4J_AUTH=neo4j/test

# install Neo4J after importing the public key from the keyserver
RUN sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 59D700E4D37F5F19 \
    && curl -fsSL https://debian.neo4j.com/neotechnology.gpg.key |sudo gpg --dearmor -o /usr/share/keyrings/neo4j.gpg \
    && echo 'deb https://debian.neo4j.com stable latest' | sudo tee -a /etc/apt/sources.list.d/neo4j.list \
    && sudo add-apt-repository universe \
    && sudo apt-get update \
    && sudo apt-get install -y neo4j
