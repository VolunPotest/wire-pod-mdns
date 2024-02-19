# Wire POD mDNS Service

This is the wire-pod mDNS instance.

The purpose of this repo to announce mDNS record of [wire-pod](https://github.com/kercre123/wire-pod) (escapepod) to vektor.

This is useful to separate logic of [wire-pod](https://github.com/kercre123/wire-pod) and mDNS broadcasting (which can be run on a different device or docker network)

## How it works

1) Check if vektor is exists in a network (with mDNS browse)
2) If vektor exists - start to announce mDNS proxy (default each 1 minute)

## Useful cases

- Separate mDNS broadcast and wire-pod instance
- Don't use host network more in wire-pod instance (see compose.yaml file)

## How to use (docker compose) - Recommended

- Download the repo
- Install docker if not installed
- Copy compose.example.yaml to compose.yaml
- Comment wire-pod service (if not needed)
- specify IP address of wire-pod instance (host IP where is the wire-pod running) in command section
- run docker compose: `docker compose up -d`

## How to use (docker)

- Download the repo
- Install docker if not installed
- build the image with command: `docker build -t "wire-pod-mdns"`
- run docker with command: `docker run -d --network host wire-pod-mdns --srvIP "provide IP of host with wire-pod running"`

## TODOs

- Make an image to docker hub
