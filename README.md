This repo contains a tool for interacting with the OpenAI API in plain text. Input is received from stdin and contains `%%`-separated messages, alternating between user and assistant roles.

```
<user message>

%%

<assistant message>

%%

<user message>

...
```

The `Chat` script is meant to be used with the [Acme](acme.cat-v.org) editor. It will read text in the current window and run a GPT chat session.