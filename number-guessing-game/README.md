# Number Guessing Game

A CLI number guessing game written in Go. The computer picks a random number between 1 and 100 — you have a limited number of attempts to guess it, depending on the difficulty level you choose.

## Project URL

https://roadmap.sh/projects/number-guessing-game

## Features

- Three difficulty levels with different attempt limits
- Remaining attempts shown after every wrong guess
- Receive a hint if you're struggling, with clues based on how close your last guess was.
- Round timer so you can see how long each round took
- View session statistics, including rounds played, rounds won, and your overall win rate.
- Best score tracking (fewest attempts to win) across the session

## Installation

Clone the repository:

```bash
git clone <repository-url>
cd number-guessing-game
```

Run without building:

```bash
go run .
```

Or build an executable:

```bash
go build -o number-guessing-game
```

Then run it:

```bash
./number-guessing-game
```

## How to Play

1. Start the game and choose a difficulty level.

   | Level  | Attempts |
   |--------|----------|
   | Easy   | 10       |
   | Medium | 5        |
   | Hard   | 3        |

2. Enter your guesses. After each wrong guess you'll see:
   - Whether the target is higher or lower than your guess
   - How many attempts you have left
   - A hint once you've used half your attempts

3. Win by guessing the number before you run out of attempts
4. After each round, choose to play again or quit to see your final summary

## Project Structure

```text
.
├── main.go    # All game logic
├── go.mod
└── README.md
```
