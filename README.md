# Nftables counter exporter

A prometheus exporter that exposes nftables counters.

## Usage

Create some counters.

```sh
sudo ./test.nft
```

Run the exporter.

```sh
sudo podman run -d --rm --name nftables-exporter --network host --cap-drop all --cap-add net_admin quay.io/itix/nftables-exporter:latest
```
