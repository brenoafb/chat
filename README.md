This repo contains a tool for interacting with the OpenAI API in plain text. Input is received from stdin and contains `%%`-separated messages, alternating between user and assistant roles.

```
<user message>

%%

<assistant message>

%%

<user message>

...
```

You can provide a system prompt by either pointing to a file with the `-s` flag or by providing it in the input -- input starting with a line containing only `%%` will be interpreted as follows:


```
```
%%

<system message>

%%

<user message>

%%

<assistant message>

%%

<user message>

...
```
```

The `Chat` script is meant to be used with the [Acme](acme.cat-v.org) editor. It will read text in the current window and run a GPT chat session.