# gator

`gator` is a command-line RSS feed aggregator. You add RSS feeds, follow the ones
you like, and let it collect new posts in the background. Then you browse the
latest posts right from your terminal. Everything is stored in PostgreSQL, and
several users can share the same database.

## Requirements

You need two things installed before you can run `gator`:

- **Go** 1.27 or newer, to build and install the CLI. Get it from
  [go.dev/dl](https://go.dev/dl/).
- **PostgreSQL** 18 or newer, to store users, feeds, and posts. The schema uses
  the built-in `uuidv7()` function, which first appeared in Postgres 18.

You will also use [goose](https://github.com/pressly/goose) once, to create the
database tables:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Install

Install the CLI with `go install`:

```bash
go install github.com/anggerpputro/go-gator-cli@latest
```

Go puts the binary in `$(go env GOPATH)/bin`, so make sure that directory is on
your `PATH`. The binary is named after the module, so the command is
`go-gator-cli`. If you'd rather type `gator`, add an alias:

```bash
alias gator="go-gator-cli"
```

The examples below use `gator`.

## Setup

### 1. Create the database

Start Postgres and create an empty database called `gator`:

```bash
psql -U postgres -c "CREATE DATABASE gator;"
```

### 2. Create the config file

`gator` reads its settings from `.gatorconfig.json` in your home directory. Create
`~/.gatorconfig.json` with your connection string:

```json
{
    "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
}
```

Change the user, password, host, and port to match your Postgres setup. The
program won't start if this file is missing.

You don't need to set `current_user_name` yourself. `gator` writes it to the same
file when you run `register` or `login`.

### 3. Run the migrations

The schema lives in `sql/schema`. Clone the repo and apply it with goose:

```bash
git clone https://github.com/anggerpputro/go-gator-cli.git
cd go-gator-cli
goose -dir sql/schema postgres "postgres://postgres:postgres@localhost:5432/gator" up
```

Use the same connection string as in your config file, without the
`?sslmode=disable` part.

## Usage

Every command follows the same shape:

```bash
gator <command> [arguments]
```

A typical first session looks like this:

```bash
# Create a user. This also logs you in.
gator register alice

# Add a feed. You automatically follow feeds you add.
gator addfeed "Boot.dev Blog" https://blog.boot.dev/index.xml

# Collect posts every minute. Leave this running in its own terminal,
# and stop it with Ctrl+C.
gator agg 1m

# In another terminal, read the 10 newest posts from feeds you follow.
gator browse 10
```

### Commands

| Command | What it does |
| --- | --- |
| `register <name>` | Creates a new user and logs in as that user. |
| `login <name>` | Switches to an existing user. |
| `users` | Lists all users and marks the one you're logged in as. |
| `addfeed <name> <url>` | Adds an RSS feed and follows it. |
| `feeds` | Lists every feed in the database and who added it. |
| `follow <url>` | Follows a feed that someone else already added. |
| `following` | Lists the feeds you follow. |
| `unfollow <url>` | Stops following a feed. |
| `agg <interval>` | Fetches feeds in a loop, one feed per tick. The interval is a Go duration such as `30s`, `1m`, or `1h`. |
| `browse [limit]` | Shows the newest posts from feeds you follow. The default limit is 2. |
| `reset` | Deletes every user, along with their feeds, follows, and posts. Useful while testing. |

`addfeed`, `follow`, `following`, `unfollow`, and `browse` need a logged-in user.
Run `register` or `login` first.

Please don't set `agg` to a very short interval. Each tick sends a request to
someone's server, so a few seconds between requests is the minimum, and a minute
or more is kinder.
