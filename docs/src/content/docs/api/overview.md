---
title: API Overview
description: Use the HTTP API from scripts after you have configured Blockbusterr in the web interface.
---

**You do not need the API to use Blockbusterr.** Connections, rules, jobs, previews, and history are available in the web interface. If you are setting up your first job, start with the [quickstart](/getting-started/quickstart/).

This section is for scripts and other applications that need to read or control Blockbusterr. Examples use `curl`, a command-line HTTP client.

## Address and authentication

The API lives under `/v1`, for example:

```text
http://localhost:9090/v1
```

The `/v1` prefix is the API path; it does not mean you are running Blockbusterr v1. Replace `localhost:9090` with the address of your instance.

When `BLOCKBUSTERR_AUTH_TOKEN` is configured, use HTTP Basic authentication with username `blockbusterr` and that token. For an interactive command, `curl --user blockbusterr` prompts for the token as a password:

```bash
curl --user blockbusterr http://localhost:9090/v1/jobs/list
```

Without owner authentication, omit `--user blockbusterr`. Keep unauthenticated instances on a trusted network, and use HTTPS when sending credentials over a remote connection.

## Find the job before calling it

```bash
curl --user blockbusterr http://localhost:9090/v1/jobs/list
```

Find the job's stable `id` in the response. Its display name is not necessarily its ID. The examples below use `JOB_ID` as a placeholder: replace it with a real ID from your installation.

### Preview without delivering

```bash
curl --user blockbusterr -X POST \
  http://localhost:9090/v1/jobs/JOB_ID/preview
```

Preview fetches candidates and evaluates rules without delivering them. It still calls the discovery source and checks destination state, so connection errors are possible. An ordinary preview's eligible count does not simulate every live delivery cap. See [Jobs](/concepts/jobs/).

### Run an enabled job

```bash
curl --user blockbusterr -X POST \
  http://localhost:9090/v1/jobs/JOB_ID/trigger
```

This starts execution asynchronously. **It can add titles or make requests** unless installation-wide dry-run is enabled. A successful trigger response means execution was started, not that every delivery succeeded. Follow the result in **Activity → Job Runs**.

Jobs participating in active ranked selection run through the shared cycle rather than this individual trigger. See the [Jobs API](/api/jobs/) for selection endpoints.

### Read recent Activity Entries

```bash
curl --user blockbusterr \
  'http://localhost:9090/v1/activity/logs?limit=50'
```

The [Activity guide](/concepts/activity/) explains what the outcomes mean; the [Activity API](/api/activity/) describes query parameters and response fields.

## Responses and errors

Most endpoints return JSON. The response shape depends on the endpoint: do not assume every response has a `data` wrapper. Use each reference page's documented fields.

| Code | Meaning | Usual next step |
| --- | --- | --- |
| 200 | Request succeeded | For asynchronous work, inspect the resulting Job Run |
| 400 | Invalid request or settings | Read the response error and correct the input |
| 401 | Authentication required or invalid | Check the owner username and token |
| 404 | Endpoint or resource not found | Check the URL and job/rule ID |
| 409 | Conflict, such as a stale revision or resource still in use | Read current state before retrying an update |
| 500 | Server-side failure | Read the error and relevant logs |

Do not blindly retry delivery-triggering requests after a timeout. The job may have started despite the lost response; check Job Runs first.

## Reference pages

- [Jobs API](/api/jobs/): job types, recipes, creation, previews, execution, and ranked selection.
- [Rules API](/api/rules/): rule sets, revisions, copies, and title exceptions.
- [Activity API](/api/activity/): history, Job Runs, and cleanup.
- [Configuration API](/api/config/): settings, connection checks, and provider metadata.

Legacy named-job endpoints remain available for compatibility. New scripts should use the dynamic job IDs and endpoint forms shown above.
