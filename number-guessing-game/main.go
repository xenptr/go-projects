package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	MinNumber = 1
	MaxNumber = 100

	Easy   = 10
	Medium = 5
	Hard   = 3
)

var scanner = bufio.NewScanner(os.Stdin)

// readLine: reads one full line from stdin, trims surrounding whitespace,
func readLine(prompt string) (string, bool) {
	fmt.Print(prompt)
	if !scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			fmt.Println("\nError reading input:", err)
		} else {
			fmt.Println("\nInput ended.")
		}
		return "", false
	}
	return strings.TrimSpace(scanner.Text()), true
}

// readInt: prints prompt and loops until the user enters a valid integer.
func readInt(prompt string) (int, bool) {
	for {
		line, ok := readLine(prompt)
		if !ok {
			return 0, false
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println("That doesn't look like a number. Please try again.")
			continue
		}
		return n, true
	}
}

func main() {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║    Welcome to Number Guessing Game!  ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("I'm thinking of a number between %d and %d.\n", MinNumber, MaxNumber)
	fmt.Println("Try to guess it within the allowed number of attempts.")
	fmt.Println()

	fmt.Println("Select a difficulty level:")
	fmt.Println("  1. Easy   — 10 attempts")
	fmt.Println("  2. Medium —  5 attempts")
	fmt.Println("  3. Hard   —  3 attempts")
	fmt.Println()

	choice := readDifficultyChoice()
	if choice == 0 {
		return
	}

	difficulty, chances := getDifficulty(choice)
	fmt.Printf("\nDifficulty set to %s (%d attempts per round). Good luck!\n", difficulty, chances)
	fmt.Println()

	// Session stats
	var (
		roundsPlayed int
		roundsWon    int
		bestAttempts int // fewest attempts to win; 0 means no wins yet
	)

	for {
		roundsPlayed++
		fmt.Printf("─── Round %d ───────────────────────────\n", roundsPlayed)

		target := rand.IntN(MaxNumber-MinNumber+1) + MinNumber
		start := time.Now()

		won, attempts := playRound(target, chances)

		elapsed := time.Since(start).Truncate(time.Millisecond)

		if won {
			roundsWon++
			fmt.Printf("⏲️  Time: %s\n", elapsed)
			if bestAttempts == 0 || attempts < bestAttempts {
				bestAttempts = attempts
				fmt.Printf("🏆 New best score: %d attempt(s)!\n", bestAttempts)
			} else {
				fmt.Printf("   Personal best: %d attempt(s).\n", bestAttempts)
			}
		} else {
			fmt.Printf("The correct number was %d. Better luck next time!\n", target)
			fmt.Printf("⏲️  Time: %s\n", elapsed)
		}

		fmt.Println()
		fmt.Printf("   Rounds played: %d\n", roundsPlayed)
		fmt.Printf("   Rounds won:    %d\n", roundsWon)
		fmt.Println()

		if !askPlayAgain() {
			break
		}
		fmt.Println()
	}

	// Final summary
	fmt.Println("════════════════════════════════════════")
	fmt.Println("           Final Summary")
	fmt.Println("════════════════════════════════════════")
	fmt.Printf("  Difficulty:    %s\n", difficulty)
	fmt.Printf("  Rounds played: %d\n", roundsPlayed)
	fmt.Printf("  Rounds won:    %d\n", roundsWon)
	if roundsPlayed > 0 {
		winRate := float64(roundsWon) / float64(roundsPlayed) * 100
		fmt.Printf("  Win rate:      %.0f%%\n", winRate)
	}
	if bestAttempts > 0 {
		fmt.Printf("  Best score:    %d attempt(s)\n", bestAttempts)
	}
	fmt.Println()
	fmt.Println("Thanks for playing — see you next time!")
}

// readDifficultyChoice: loops until the user enters 1, 2, or 3.
// Returns 0 on input error.
func readDifficultyChoice() int {
	for {
		n, ok := readInt("Enter your choice (1/2/3): ")
		if !ok {
			return 0
		}
		if n >= 1 && n <= 3 {
			return n
		}
		fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
	}
}

func getDifficulty(choice int) (string, int) {
	switch choice {
	case 1:
		return "Easy", Easy
	case 2:
		return "Medium", Medium
	case 3:
		return "Hard", Hard
	default:
		return "", 0
	}
}

// playRound: runs a single round. Returns (won, attemptsTaken).
func playRound(target, chances int) (bool, int) {
	hintGiven := false
	hintThreshold := max(chances/2, 1)

	for attempt := 1; attempt <= chances; attempt++ {
		guess, ok := readInt(fmt.Sprintf("Attempt %d/%d — Enter your guess: ", attempt, chances))
		if !ok {
			return false, 0
		}

		if guess < MinNumber || guess > MaxNumber {
			fmt.Printf("Please enter a number between %d and %d.\n\n", MinNumber, MaxNumber)
			attempt--
			continue
		}

		switch {
		case guess > target:
			fmt.Printf("Too high! The number is less than %d.\n", guess)
		case guess < target:
			fmt.Printf("Too low!  The number is greater than %d.\n", guess)
		default:
			fmt.Printf("\n🎉 Correct! You guessed %d in %d attempt(s).\n", target, attempt)
			return true, attempt
		}

		// Show remaining attempts
		remaining := chances - attempt
		if remaining > 0 {
			if remaining == 1 {
				fmt.Printf("⚠️  Last attempt remaining!\n")
			} else {
				fmt.Printf("   %d attempt(s) remaining.\n", remaining)
			}
		}

		// Offer a hint once when the threshold is crossed
		if !hintGiven && attempt >= hintThreshold {
			hintGiven = true
			showHint(target, guess)
		}

		fmt.Println()
	}

	return false, 0
}

func showHint(target, lastGuess int) {
	diff := target - lastGuess
	if diff < 0 {
		diff = -diff
	}

	fmt.Print("💡 Hint: ")
	switch {
	case diff <= 5:
		fmt.Println("You're very close — within 5!")
	case diff <= 15:
		fmt.Println("Getting warmer — within 15.")
	case target%10 == 0:
		fmt.Println("The number is a multiple of 10.")
	case target%5 == 0:
		fmt.Println("The number is a multiple of 5.")
	case target%3 == 0:
		fmt.Println("The number is divisible by 3.")
	case target%2 == 0:
		fmt.Println("The number is even.")
	default:
		fmt.Println("The number is odd.")
	}
}

func askPlayAgain() bool {
	for {
		line, ok := readLine("Play again? (y/n): ")
		if !ok {
			return false
		}
		switch strings.ToLower(line) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please enter y or n.")
		}
	}
}
