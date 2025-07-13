# LLM Chat CLI

**LLMChatCLI** is a simple command-line tool for interacting with different LLM models and providers that use OpenAI-compatible API. It lets you load structured conversation inputs, engage in chat sessions, and automatically log responses for later analysis.

This tool is designed for testing and comparing outputs from different LLMs using consistent conversation history and prompts.


---

## Features

- Chat with any OpenAI-compatible LLM provider via CLI.
- Load messages and system prompts from structured JSON or text files.
- Easily switch models or providers via flags or environment variables.
- Save complete chat logs automatically for future inspection.
- Supports ongoing conversations with pre-defined chat history.

---

## Setup

1. **Clone the Repository**

  ```bash
  git clone git@github.com:igorMSoares/llm-chat-cli.git
  cd llm-chat-cli
  ```

2.  **Build the application**

    This command will also automatically download and install the necessary dependencies.

    ```bash
    go build -o llm-chat-cli
    ```

    _Go must be installed: [Install Go](https://go.dev/doc/install)_

3.  **Create a `.env` file**

    Copy the `.env.example` file to a new file named `.env`:

    ```bash
    cp .env.example .env
    ```

    Then, open the `.env` file and fill in the required environment variables:

    *   `LLM_PROVIDER_KEY`: Your API key for the LLM provider.
    *   `LLM_MODEL`: The name of the LLM model you want to use.
    *   `CHAT_COMPLETION_URL`: The URL for the chat completion API (must be OpenAI compatible).
    *   `TEMPERATURE`: The temperature for the LLM (optional, defaults to 0).

## Usage

After building the application, you can run it from the terminal:

```bash
./llm-chat-cli [flags]
```

On Windows, the command would be:

```bash
.\llm-chat-cli.exe [flags]
```

### Available Flags

The application supports the following command-line flags:

| Flag            | Description                                                       |
| --------------- | ----------------------------------------------------------------- |
| `--api-key`     | API key for the LLM provider (overrides `LLM_PROVIDER_KEY`)       |
| `--model`       | Name of the LLM model to use (overrides `LLM_MODEL`)              |
| `--url`         | Chat API URL (overrides `CHAT_COMPLETION_URL`)                    |
| `--temperature` | Sampling temperature (overrides `TEMPERATURE`)                    |
| `--input`       | Input file name (default: `messages.json`)                        |
| `--input-dir`   | Directory containing input files (default: `input`)               |
| `--prompts-dir` | Directory containing prompt files (default: `prompts`)            |
| `--logs-dir`    | Directory where conversation logs will be saved (default: `logs`) |

#### Example

To run using the example messages:

```bash
./llm-chat-cli --input messages.example.json
```

### Input File

The input file is a JSON file that contains an array of messages, which can be used to set the context for the conversation or to load an ongoing chat history.

Each message is an object with the following properties:

*   `role`: The role of the message sender. Can be `user`, `assistant`, or `system`.
*   `content`: The content of the message.
*   `file`: (Optional) The name of a file containing the system message. This is only used for `system` messages and will be loaded from the directory specified by `--prompts-dir`.

_The input file can not be empty. It must be valid JSON and contain at least an empty array: `[]`._

#### Behavior on Startup

The application's initial behavior depends on the role of the *last* message in the input file:

*   If the last message is from the `system` or `assistant`, the application will prompt you for input to start the conversation.
*   If the last message is from the `user`, the application will immediately send the entire conversation history to the LLM, display the assistant's response, and then prompt you for your next message.

#### Examples

**1. Starting with a system prompt:**

In this example, the system message is read from the `prompts/system_prompt.md` file. The application will start by asking for user input because the last message is not from the user.

```json
[
  {
    "role": "system",
    "file": "system_prompt.md"
  }
]
```

**2. Loading a conversation history:**

This example loads a pre-existing conversation. Since the last message is from the `user`, the application will start by sending the request to the API to get the assistant's response before prompting for the next user input.

```json
[
  {
    "role": "system",
    "content": "You are a helpful assistant."
  },
  {
    "role": "user",
    "content": "Hello, who are you?"
  },
  {
    "role": "assistant",
    "content": "I am a helpful assistant. How can I help you today?"
  },
  {
    "role": "user",
    "content": "What is the capital of Butan?"
  }
]
```

### Commands

While chatting with the model, you can use the following commands:

| Command  | Description                                      |
| -------- | ------------------------------------------------ |
| `/quit`  | Save the conversation log and exit               |
| `/quit!` | Exit immediately without saving the conversation |

## Contributing

Contributions are welcome and encouraged!

### Ways to contribute:

-   Open issues for bugs, ideas, or questions
    
-   Submit pull requests to fix, improve, or add new features
    
-   Improve documentation or add examples
    

### To get started:

1.  Fork the repository
    
2.  Create a feature branch (`git checkout -b my-feature`)
    
3.  Make your changes
    
4.  Open a pull request with a clear description
    

If you're unsure where to start, feel free to open an issue or ask!

----------

## License

This project is licensed under the MIT License. You are free to use, modify, and distribute it.

----------

## Show Your Support

If you find this project helpful, consider starring the repo to show your support!
