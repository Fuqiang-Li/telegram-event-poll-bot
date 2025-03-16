# telegram-event-poll-bot

## Overview
`telegram-event-poll-bot` is a Telegram bot designed to create and manage event polls. Users can create events, set various parameters, and allow participants to vote on different options.

## Features
- Create events with descriptions, start times, desired and maximum participants, and options.
- Send event polls to groups.
- Collect and display votes from participants.

## Installation
1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/telegram-event-poll-bot.git
    cd telegram-event-poll-bot
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```
3. Create a `config.json` file by copying the config_template.json files and update your Telegram bot token

## Usage
1. Run the bot:
    ```sh
    go run .
    ```

2. Interact with the bot on Telegram:
    - Use `/start` to begin creating an event.
    - Follow the prompts to set event details.
    - Use `/send` to send the event poll to a group.

## File Structure
- `main.go`: Entry point of the application. Initializes the bot and sets up handlers.
- `createEventHandler.go`: Handles the creation of events and user interactions.
- `event.go`: Defines the `Event` and `EventAndUsers` structs and related methods.
- `utils.go`: Utility functions for common tasks.
