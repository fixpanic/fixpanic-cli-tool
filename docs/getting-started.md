# Getting Started

Follow these steps to deploy your first node.

## 1. Install a Node
To install and register a node, you need a **Node ID** and **Token** from the OpsSquad dashboard.

Run the following command:

```bash
opssquad node install \
  --node-id="your-node-id" \
  --token="your-token"
```

This command will:
- Register the node with the OpsSquad platform.
- Create the necessary configuration files.
- Start the node service.

## 2. Check Status
Verify that your node is running and connected:

```bash
opssquad node status
```

You should see output indicating the node is **Active** and **Connected**.

## 3. View Logs
To monitor the node's activity in real-time:

```bash
opssquad node logs --follow
```

This is useful for verifying that the node is receiving tasks and executing them correctly.
