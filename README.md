# CreemProxy

A proxy designed for [Creem](https://creem.io) users to securely delegate requests to the Creem API. It also signs responses with a private key, enabling client-side verification.

## Features

- **Proxying**: Protects your API key by proxying requests to the upstream API.
- **Response Signing**: Signs responses with a private key to ensure data integrity and authenticity.
- **SSL/TLS**: Enables secure communication with SSL/TLS.

## Getting Started - Server Side

The simplest way to run the proxy is with Docker. Ensure you have Docker with Docker Compose support installed. Then, run `docker compose up` using the following `compose.yml` configuration:

```yaml
services:
  app:
    image: ghcr.io/lakr233/creemproxy:main
    ports:
      - "8443:8443"
    environment:
      - CREEM_API_KEY="" # remember to replace with your API key
    volumes:
      - ./data:/app/data
    restart: always
    logging:
      driver: "json-file"
      options:
        max-size: "128m"
        max-file: "5"
```

Configure the proxy using the following environment variables. These can be set in your `compose.yml` file or an `.env` file.

| Variable Name           | Description                             | Example/Default Value            |
| ----------------------- | --------------------------------------- | -------------------------------- |
| `CREEM_API_KEY`         | Header `x-api-key`                      | N/A                              |
| `CREEM_API_HOST`        | Upstream endpoint to be used            | `https://api.creem.io` (default) |
| `SERVER_LISTEN_ADDRESS` | Listen address                          | `0.0.0.0` (default)              |
| `SERVER_LISTEN_PORT`    | Listen port                             | `8443` (default)                 |
| `SERVER_DATA_DIR`       | Directory to store generated certificates and keys | `/app/data/` (default)           |

To use Creem Testing Mode, set `CREEM_API_HOST` to `https://test-api.creem.io`.

## Getting Started - Client Side

After starting the server, check its console output with `docker compose logs` for the following keys before configuring your client:

```
✔ Container creemproxy-app-1
...
app-1  |
app-1  | 2025/05/09 17:58:38 [+] certificate fingerprint (sha1): FDFA378E65E06CA4F8CCF397AA1C1148811C3CA3
app-1  | 2025/05/09 17:58:38 [+] signing public key (base64): 1/UKeIXpIeE6kbsFeTvtgxOIkkaB7n/2YMpdZx9XNCs=
...
```

The certificate fingerprint is the SHA1 hash of the SSL public key. The signing public key, encoded in Base64, is used to sign response data and utilizes Curve25519.

### DIY your own client

On the client-side, you need to:

- Handle the self-signed certificate for network communication.
- Verify the response data signature to ensure its authenticity.

The second step (signature verification) is highly recommended. The first step (handling the self-signed certificate) can be accomplished in two ways:

- Accept any certificate (less secure).
- Pin the self-signed certificate (more secure, recommended).

Pinning the certificate requires updating your client if the server's certificate changes. If this maintenance is not feasible, you might consider accepting any certificate, though this approach is less secure.

**It is strongly recommended to implement at least one of these security measures. If you opt to accept any certificate, ensure you still verify the response data signature.**

### CreemKit - Swift

For Swift developers, the [CreemKit](https://github.com/Lakr233/CreemKit) library simplifies integration with the proxy.

```swift
import CreemProxyKit

let creemInterface = CreemInterfaceViaProxy(
    host: try! String(
        contentsOfFile: "/tmp/creem.api.test.host"
    ).trimmingCharacters(in: .whitespacesAndNewlines),
    certificateFingerprint: .matchHash("0EA6CC5707C4485FF5A93F2D452B1903E3E6773B"),
    signingPublicKey: .verifyWithPublicKey("ObReLUe4wm8rwkBF6L99yxvm8OFrV4LADID61tpMjig=")
)
```

Once the interface is created, you can use it to `activate`, `validate`, or `deactivate` licenses. For more details, refer to the [CreemKit](https://github.com/Lakr233/CreemKit) documentation. CreemKit also offers `CreemInterface` for direct upstream connections, suitable for trusted environments like your own server (e.g., when using Vapor).

## License

MIT License. See [LICENSE](LICENSE) for details.

---

Copyright © 2025 Lakr Aream. All Rights Reserved.
