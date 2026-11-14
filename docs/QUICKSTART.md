# Quick start

This takes a few minutes and needs no API keys. Every command and reply below was run
against the code as it stands.

## 1. What you need

Go 1.26 or newer (`go version`). Herald uses the standard library only, so there is
nothing else to install.

## 2. Get the code and check it

```bash
git clone https://github.com/MaxsysLab/herald
cd herald
make check     # gofmt, go vet, go test
```

All tests should pass.

## 3. Run the offline demo

```bash
make demo
```

The demo sets up a sample café and asks five questions:

```
Q: How much is a cappuccino?
  [answer] Cappuccino is USD 4.50.
    ↳ cite profile:pricing (1.00)

Q: What are your hours?
  [answer] Bloom Café hours: Mon-Fri 07:00–18:00; Sat-Sun 08:00–16:00.
    ↳ cite profile:hours (1.00)

Q: Do you have wifi?
  [answer] Yes! Free guest wifi is available ...
    ↳ cite kb:amenities.md (0.35)

Q: How much does catering cost?
  [escalate] That's a great question ... let me connect you with someone from Bloom Café who can help.
    ↳ handoff → manager@bloomcafe.example (price_not_listed)

Q: What's the meaning of life?
  [escalate] ...
    ↳ handoff → manager@bloomcafe.example (ungrounded)
```

That is the whole idea in one screen. Exact facts come from the business profile. Open
questions are answered only when the knowledge base supports them. Everything else goes
to a person.

## 4. Start the server

```bash
go run ./cmd/herald serve -addr 127.0.0.1:8080 -config my-business.local.json
```

- `-addr 127.0.0.1:8080` keeps the server on your own machine. Without `-addr` it
  listens on port 8080 on every network interface, and the console has no sign-in, so
  use the `127.0.0.1` form unless you have read [DEPLOYMENT.md](DEPLOYMENT.md).
- `-config` names the configuration file. If it does not exist, Herald creates it from
  the sample café so you have something to edit. Files ending in `.local.json` are
  ignored by git, so your business data stays out of the repository.

## 5. Open the console

Go to `http://localhost:8080/admin`. You will see:

- **Status:** whether an LLM key was found, which embedder is in use, how many
  knowledge passages are loaded, and the grounding threshold.
- **Test:** a chat box wired to the same officer your customers would talk to.
- **Add knowledge:** paste a fact, an FAQ answer or a policy, give it a source label,
  and it is indexed at once.
- **Configuration (JSON):** the whole configuration, editable. Saving it rebuilds the
  officer and writes the file.
- **Embed on your website:** the one-line script tag.

## 6. Ask it questions

In the console's test box, or from a terminal:

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H 'content-type: application/json' \
  -d '{"message":"How much is a cold brew?"}'
```

```json
{"verdict":"answer","text":"Cold Brew is USD 5.25.","citations":[{"source":"profile:pricing","quote":"Cold Brew: USD 5.25","score":1}]}
```

A question the knowledge base cannot support is handed off:

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H 'content-type: application/json' \
  -d '{"message":"Do you sell gift cards?"}'
```

```json
{"verdict":"escalate","text":"That's a great question ... let me connect you with someone from Bloom Café who can help.","escalation":{"reason":"ungrounded","contact":"manager@bloomcafe.example","summary":"Customer asked: Do you sell gift cards?"}}
```

Now add the missing fact in the console ("gift cards are sold at the counter in any
amount") and ask again. The officer answers from it, and cites the source you gave.

Every reply carries one of three verdicts:

| Verdict | Meaning |
|---|---|
| `answer` | Backed by a profile fact or a knowledge passage; citations are included |
| `escalate` | In scope but not supported; the reply offers a person, and `escalation` says why |
| `refuse` | Empty or unsafe input; declined politely |

## 7. Make it yours

Edit the configuration in the console or in the file: your business name, services and
prices, hours, policies, and an escalation contact. [CONFIGURATION.md](CONFIGURATION.md)
explains every field.

## 8. Turn on natural drafting (optional)

Without a key, Herald quotes the best-matching knowledge passage word for word. With
an LLM key it phrases the answer naturally, still held to the retrieved passages.

```bash
export GROQ_API_KEY=...        # or OPENAI_API_KEY, XAI_API_KEY; see CONFIGURATION.md
go run ./cmd/herald serve -addr 127.0.0.1:8080 -config my-business.local.json
```

The console's status line changes from "verbatim (no key)" to "drafting". Before you
do this, read [DATA.md](DATA.md): with a key set, customer questions are sent to that
provider.

## 9. Put the chat on a website

Add this line before `</body>` on any page, with the address where your Herald runs:

```html
<script src="https://YOUR-HERALD-HOST/widget.js" defer></script>
```

A chat bubble appears in the corner. For a real site, follow
[DEPLOYMENT.md](DEPLOYMENT.md) first.
