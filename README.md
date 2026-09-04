# Go Projects by roadmap.sh

This repository contains the projects I build while learning backend development with Go.

Projects are organized into separate directories, each with its own source code and documentation.

## Beginner Projects

1. **Task Tracker**
   A CLI task manager with JSON persistence. Add, update, delete, and filter tasks by status.
   Project URL: https://roadmap.sh/projects/task-tracker

2. **GitHub Activity**
   A CLI tool that fetches and displays the recent public activity of any GitHub user via the GitHub Events API.
   Project URL: https://roadmap.sh/projects/github-user-activity

3. **Expense Tracker**
   A CLI tool to manage personal finances. Add expenses, filter by category, set monthly budgets, and export to CSV.
   Project URL: https://roadmap.sh/projects/expense-tracker

4. **Number Guessing Game**
   A CLI game where the computer picks a number between 1 and 100 and you have to guess it. Choose a difficulty level, get proximity hints, and track your best score across multiple rounds.
   Project URL: https://roadmap.sh/projects/number-guessing-game

5. **Unit Converter**
   A web-based unit converter served by Go's standard library. Converts between units of length, weight, and temperature through a browser UI with no external dependencies.
   Project URL: https://roadmap.sh/projects/unit-converter

6. **Personal Blog**
   A web-based personal blog with a public reading section and a password-protected admin dashboard to create, edit, and delete articles. Built entirely with Go's standard library, no external dependencies.
   Project URL: https://roadmap.sh/projects/personal-blog

7. **Weather API**
   An HTTP API that returns weather data for a given city using the Visual Crossing API. Responses are cached in Redis for 12 hours and there's a per-IP rate limiter built in.
   Project URL: https://roadmap.sh/projects/weather-api-wrapper-service

8. **Blogging Platform API**
   A RESTful JSON API for managing blog posts backed by PostgreSQL. Full CRUD on posts plus a keyword search across title, content, category, and tags.
   Project URL: https://roadmap.sh/projects/blogging-platform-api

9. **Todo List API**
   A RESTful API for managing personal to-do lists. Includes user registration and login, JWT access tokens with refresh token rotation, full CRUD on todos with ownership checks, pagination, search, and per-IP rate limiting via Redis.
   Project URL: https://roadmap.sh/projects/todo-list-api

10. **Expense Tracker API**
   A RESTful API for managing personal expenses. Includes user signup and login, JWT access tokens with refresh token rotation, full CRUD on expenses with ownership checks, date-range filtering (past week, month, 3 months, or custom), category validation, and Redis-backed rate limiting.
   Project URL: https://roadmap.sh/projects/expense-tracker-api

11. **TMDB CLI**
    A command line tool to fetch and display popular, top-rated, upcoming, and now playing movies from The Movie Database (TMDB) API.
    Project URL: https://roadmap.sh/projects/tmdb-cli

12. **Caching Proxy**
    A caching reverse proxy CLI that forwards HTTP requests to an origin server and caches responses in Redis. Sets `X-Cache: HIT/MISS` headers and supports a `--clear-cache` flag to flush the cache.
    Project URL: https://roadmap.sh/projects/caching-server

## More Projects

Additional projects will be added as I progress through the roadmap.
