# Long Running Connections on Kubernetes

## Introduction

One of the main issues we have on Kubernetes Networking is handling long-running connections.
These connections may get disconnected, but it takes time (based on the OS TCP configuration) to figure out,
so we need an application-layer way to detect and react to these issues.

Because the solution should be in the application layer, we need to consider different ways for different
protocols.

## Deleting by Force

Forcefully deleting a Kubernetes pod (e.g., using `kubectl delete pod <pod-name> --force --grace-period=0`) **can definitely cause a delay in the *client-side* closing of a long-running TCP connection.**

Here's a breakdown of why:

1.  **Normal (Graceful) Termination:**
    * When you delete a pod normally, Kubernetes sends a `SIGTERM` signal to the container's main process.
    * The application running inside the pod is expected to catch this signal and initiate a graceful shutdown.
    * For a TCP server, this typically involves:
        * Stopping acceptance of new connections.
        * Processing any ongoing requests.
        * Properly closing existing TCP connections by initiating the standard TCP FIN/ACK handshake with each connected client.
    * The client receives the FIN packet, acknowledges it, and closes its end of the connection relatively quickly and cleanly.
    * If the application doesn't shut down within the `terminationGracePeriodSeconds` (default 30s), Kubernetes sends `SIGKILL`, forcefully terminating it.

2.  **Forceful Termination (`--force --grace-period=0`):**
    * This bypasses the `SIGTERM` signal and the grace period entirely.
    * Kubernetes immediately sends `SIGKILL` to the process inside the container.
    * `SIGKILL` cannot be caught or handled by the application; the process is terminated instantly by the operating system kernel.
    * Crucially, the application gets **no chance** to perform the graceful shutdown steps, including sending the FIN packet to close the TCP connection properly.

3.  **Impact on the TCP Connection (Client-Side):**
    * **Server-Side:** The OS on the node where the pod was running cleans up the resources associated with the killed process, including the TCP socket state. The connection is effectively gone from the server's perspective.
    * **Client-Side:** The client is unaware that the server process has vanished instantly.
        * **If the client tries to send data:** It will likely receive a TCP Reset (RST) packet back from the server's node (or an intermediary like a load balancer), as the kernel knows the destination socket is no longer valid. Receiving an RST causes the client OS to immediately close the connection and usually results in an error like "Connection reset by peer" in the client application. This is relatively quick but abrupt.
        * **If the client is idle (waiting to receive data):** This is where the delay happens. The client will simply continue waiting. The connection will only be detected as dead when:
            * The client eventually tries to send data (triggering the RST as above).
            * TCP Keep-Alives (if enabled and configured on *both* client and server sides, and not blocked by firewalls) eventually time out. Keep-alive probes are sent periodically on idle connections. If enough probes go unanswered, the client OS will time out the connection. Default keep-alive timeouts can be very long (e.g., 2 hours).

### Conclusion

Forcefully deleting a pod prevents the server application from initiating the standard TCP connection closure. While the server-side resources are cleaned up quickly, the client might not realize the connection is dead for a potentially long time, especially if it's idle and relying on default TCP keep-alive settings. This leads to a delay in the client-side connection closing and can leave the client application hanging or holding onto resources unnecessarily.

### Recommendation

Always favor graceful termination. Allow pods sufficient `terminationGracePeriodSeconds` to shut down cleanly, close connections properly, and ensure your application handles `SIGTERM` appropriately. Forceful deletion should only be used as a last resort for stuck pods.

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
in the production environment this may happen in one of the intermediate nodes which doesn't generate any
_FIN_ packet.

## MQTT
