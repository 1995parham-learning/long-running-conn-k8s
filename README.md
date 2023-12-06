# Long Running Connections on Kubernetes

## Introduction

One of the main issues we have on Kubernetes Networking is handling long-running connections.
These connections may get disconnected, but it takes time (based on the OS TCP configuration) to figure out,
so we need an application-layer way to detect and react to these issues.

Because the solution should be in the application layer, we need to consider different ways for different
protocols.

## NATS

NATS uses a long-running TCP connection for publishing and subscribing. It has a Ping/Pong method that we can use to detect connection failure before
getting the failure from the operating system.

For a better insight into the problem, I will demonstrate it using a simple NATS publisher. I deployed a NATS cluster on Minikube using
port-forward to connect to it and then close the port-forward command to see how it goes.

Based on the following logs, you can see it figured out instantly:

```
2023/12/03 07:42:48 Published message: hello
2023/12/03 07:42:49 Published message: hello
2023/12/03 07:42:50 Published message: hello
2023/12/03 07:42:51 Disconnected
2023/12/03 07:42:51 Not connected. Waiting to reestablish the connection.
2023/12/03 07:42:52 Not connected. Waiting to reestablish the connection.
2023/12/03 07:42:53 Not connected. Waiting to reestablish the connection.
2023/12/03 07:42:54 Not connected. Waiting to reestablish the connection.
```

This happens because I terminate connection in one of its ends and it generates the _FIN_ packet,
in the production environment this may happen in one of the intermediate nodes which don't generate any
_FIN_ packet.

## MQTT
