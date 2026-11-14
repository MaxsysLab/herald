# Data: what Herald handles and what leaves the machine

Written for the business that runs Herald, so it can tell its own customers the truth.
It describes the code as it stands. It is not legal advice; the business running
Herald is responsible for its own privacy notice and for the laws that apply to it.

## What Herald keeps

| Data | Where | For how long |
|---|---|---|
| The business's configuration: profile, prices, hours, policies, contacts, knowledge passages | One JSON file on the server, permissions `0600` | Until you change or delete it |
| The knowledge index | In memory | Until the program stops; rebuilt at start |
| Customer questions and the officer's replies | In memory, for the length of one request | Not stored. Herald writes no conversation log |

Herald sets no cookies and keeps no visitor identifiers. The widget keeps the visible
conversation in the visitor's browser page only; it is gone when the page is closed.

## What leaves the machine

That depends on which providers the business configures.

**With no keys set (the default):** nothing. Facts come from the profile, knowledge is
matched on the machine, and replies quote the knowledge word for word.

**With an LLM provider and its key set:** when a question is answered from the
knowledge base, Herald sends that provider:

- the customer's question,
- the knowledge passages that matched,
- the business's name, the officer's name, and the brand-voice line.

Questions answered from the profile (prices, hours, policies) and questions that are
handed off are **not** sent to the LLM provider.

**With a hosted embedder set (`openai`, `voyage`, or a custom endpoint):**

- every knowledge passage is sent to that provider when the index is built, at each
  start and after each change;
- every customer question that reaches the knowledge step is sent to that provider,
  to be matched against the knowledge.

Each provider has its own terms on retention and use. Read them before you choose,
and prefer a provider or a local endpoint that fits what you promise your customers.
`ollama` and a `custom` endpoint on your own machine keep everything local.

## What customers see

- Citations: the source label and the text of the passage or fact used.
- On a handoff: the escalation contact from the profile. Treat it as public.

## What the business should tell its visitors

At the least: that they are talking to an automated assistant, and where their
messages go if an LLM or embedding provider is configured. The greeting names the
officer and calls it "the assistant" for the business; whether that wording is enough
where you operate is for you to check.
