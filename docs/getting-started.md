# Getting Started

Follow these steps to deploy your first agent.

## 1. Install an Agent
To install and register an agent, you need an **Agent ID** and **API Key** from the FixPanic dashboard.

Run the following command:

```bash
fixpanic agent install \
  --agent-id="your-agent-id" \
  --api-key="your-api-key"
```

This command will:
- Register the agent with the FixPanic platform.
- Create the necessary configuration files.
- Start the agent service.

## 2. Check Status
Verify that your agent is running and connected:

```bash
fixpanic agent status
```

You should see output indicating the agent is **Active** and **Connected**.

## 3. View Logs
To monitor the agent's activity in real-time:

```bash
fixpanic agent logs --follow
```

This is useful for verifying that the agent is receiving tasks and executing them correctly.
