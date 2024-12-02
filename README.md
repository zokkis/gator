# Gator

Gator is a command-line application for managing RSS feeds and users. It allows users to register, login, add feeds, follow/unfollow feeds, and browse posts from followed feeds.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)
- [Configuration](#configuration)
- [Database Schema](#database-schema)

## Installation

1. Clone the repository:
	```sh
	git clone https://github.com/zokkis/gator.git
	cd gator
	```

2. Install dependencies:
	```sh
	go mod download
	```

3. Create DB:
	```sh
	cd ./sql/schema
	goose postgres "postgres://postgres:postgres@localhost:5432/gator" up
	cd ../..
	```

4. Create structs:
	```sh
	sqlc generate
	```

5. Build the project:
	```sh
	go build -o gator
	```

6. Install the project:
	```sh
	go install
	```

## Usage

Run the application with the desired command:
```sh
./gator <command> [args...]
```

## Commands

- `register <name>`: Register a new user with the given name.
- `login <name>`: Login as the user with the given name.
- `users`: List all registered users.
- `reset`: Delete all users from the database.
- `addfeed <name> <url>`: Add a new feed with the given name and URL.
- `feeds`: List all feeds.
- `follow <url>`: Follow the feed with the given URL.
- `following`: List all feeds the current user is following.
- `unfollow <url>`: Unfollow the feed with the given URL.
- `agg <time_between_reqs>`: Fetch feeds at the specified interval.
- `browse [limit]`: Browse posts from followed feeds, optionally limiting the number of posts.

## Configuration

The configuration file is located at `~/.gatorconfig.json`. It contains the following fields:

- `db_url`: The URL of the PostgreSQL database.
- `current_user_name`: The name of the currently logged-in user.

Example configuration:
```json
{
	"db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
	"current_user_name": "kahya"
}
```

## Database Schema

The database schema is managed using Goose migrations. The schema files are located in the `sql/schema` directory.

- `001_users.sql`: Creates the users table.
- `002_feeds.sql`: Creates the feeds table.
- `003_feed_follows.sql`: Creates the feed_follows table.
- `004_feeds_fetched_at.sql`: Adds the last_fetched_at column to the feeds table.
- `005_posts.sql`: Creates the posts table.