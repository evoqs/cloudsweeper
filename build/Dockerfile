# Use the official Ubuntu base image
FROM ubuntu:latest

# Define the username as a global constant
ENV USERNAME=csw

# Install necessary dependencies
RUN apt-get update && apt-get install -y \
    wget \
    curl \
    git \
    vim \
    net-tools \
    && rm -rf /var/lib/apt/lists/*

# Install python3
RUN apt update -y && apt install -y python3 && python3 --version && apt install -y python3-venv

# Download and install GoLang
RUN wget https://golang.org/dl/go1.22.0.linux-amd64.tar.gz
RUN tar -xvf go1.22.0.linux-amd64.tar.gz
RUN mv go /usr/local && rm -f go1.22.0.linux-amd64.tar.gz

# Create the user and set permissions
#RUN useradd -m $USERNAME
RUN useradd -s /bin/bash -m csw && usermod -aG csw csw 

# Set Go environment variables
ENV GOROOT=/usr/local/go
ENV GOPATH=$HOME/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Check if Go is installed properly
RUN go version

# (Optional) Set working directory
WORKDIR /app

# Install Cloud Custodian
RUN python3 -m venv custodian && \
    . custodian/bin/activate && \
    pip install c7n

# Clone the application
COPY ./cloudsweep/go/cloudsweep /app
COPY ./cloudsweep/go/config.yml /app
COPY ./logo.jpeg /app

RUN chown -R csw:csw /app

# Set all the env values
ENV C7N_AWS_INSTALL="/app/custodian/bin/activate"
ENV C7N_AWS_INSTALL="/app/logo.jpeg"

# Run as user
USER $USERNAME


# (Optional) Command to run your application
 CMD ["./cloudsweep"]

