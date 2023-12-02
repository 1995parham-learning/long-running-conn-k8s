# Long Running Connections on Kubernetes

## Introduction

One of the main issues we have on Kubernetes Networking is handling long-running connections.
These connections may got disconnected but it takes time (based on the OS TCP configuration) to figured out,
so we need to have application-layer way to detect these issues and react on them.

Because solution should be in the application layer, we need to consider different way for different
protocols.

## NATS

NATS uses a long running TCP connection for publishing and subscribing. It has Ping/Pong method that we can use to detect connection failure before
getting the failure from operating system.

## MQTT
