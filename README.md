# Go for Pomodoro

Simple README for a simple bot application written in Go.

Pomodoro Technique is a technique of studying or other concentration tasks
where the time is allocated during a session in more sprints, and among each
pair of sprints there is a break. E.g., 4 pomodoros (4 sprints), 25 minutes
for each pomodoro (task time), 5 minutes of rest, 25 minutes of task time, etc.
for 4 times.

The application implements a Pomodoro timer flexible to different
configurations that can be used by the users as a Telegram bot.

### Official Bot

#### How to find and use

You can find it on [this link](https://t.me/go4pom_bot). Type `/help` to learn how
the bot is used.

* `/30` will run a single Pomodoro of 30 minutes.

* `/25` will run a single Pomodoro of 25 minutes.

* `/25for4` will run a session of 4 Pomodoros of 25 minutes. The rest time by
default is set to 5 minutes.

To set a different rest time, the query is modified like

* `/25for4rest7` (now the rest time among sprints will be of 7 minutes).

By default, setting a configuration will trigger the timer to start. You can
modify this behavior by typing

`/autorun off`

In that case, the Pomodoro is started doing `/start_sprint` or just `/s`.

You can cancel a session with `/cancel` command (will make it unrestorable)
or you can temporarily `/pause` it and `/resume` it in another moment.

You can reset all the configuration associated with your chat with `/reset`.
(This operation is irreversible.)

#### Commands' groups

This bot also works in groups. In groups, you have another pair of commands
that are possible useful, and they are

* `/join`
* `/leave`

After you `/join` the bot in a group, the bot will tag you by username each
update of a session (session start, break, resume, finish, etc.), so that you
get notifications in the group chat even if it was otherwise silenced.

Vice versa, with `/leave` you signal that you wish not to be notified anymore
for the updates. You may always join later.

Remind that `/reset` also works in group and will also un-join all the chat
members.

## Licensing

GNU AGPL 3 (Affero General Public License), since this application is not a
library, and the purpose of this project and alike believes that not only
open-source is the most appropriate form to publish software, but also that
the users deserve to have access to the code of the software they're using.

The GNU GPL-3 has **not** be used in this project as it would not be very much
meaningful for the purpose. The application is mainly server-side-like, and
offers a service to the users. The GNU GPL-3 enforces the distribution of the
code alongside the binary but not the distribution of the code in
Software-as-a-Service model.

Since this application offers a SaaS, and I want it open-source also in its
re-distributions, the AGPL license is the fittest license.

## Development stack

- Go (1.19, need Generics to work)
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- SQLite ([driver](modernc.org/sqlite))
- not very much else

### Why SQLite?

SQLite is a simple and very powerful DBMS, often under-estimated among
developers.

Free and open-source, easy to configure and use, very good performances and
efficiency.

I am aware that possibly it isn't the most efficient DB for an application
like this. For this reason, although not exclusively, the DB side is very much
abstracted in the program. The components that touch the DB (in `data` package)
do not use SQL or SQLite directly; instead, they refer to an abstract
key-value-store. Such key-value-store as of now has SQLite as backend, but
it would be really easy to implement another backend (e.g., using Redis
instead) under the same interface and providing it in the place of
`persistence.Manager` interface (dependency injection pattern is used here not
to force a particular DB onto the application).

This software is Free and Open-Source and as such, you're free to implement
your own a different DB underneath and eventually to make a pull request for
its integration. Any contributions to this project would be appreciated.

### Why Go?

A lot of Telegram bots are often written in either JS or Python. Go is no less
safe than these languages, but allows for a very more robust concurrency model
and better performance. Since a bot can have a lot of users at the same
time, these advantages are well appreciated. At the same time, Go doesn't
constitute an obstacle to what the project's purpose is, and I hope that
readability is also a good side of this choice and the code itself.

## How to run the bot

### Getting a token
The application needs a proper authentication with Telegram to work. It's
assumed that who wants to run this bot has a valid Telegram account and at
least a bot key available to use.

If you have a Telegram account, you can create a bot using @BotFather, I
recommend the [official guide](https://core.telegram.org/bots#6-botfather).

### Setting the token for the application

Create a file named `appsettings.toml` in the directory the application will
be run. (It can also be the project directory), and type inside

```toml
ApiToken = "<your api token>"
BotName = "@<your bot username>"
```

### Testing that the configuration is OK

You can test the configuration by running

```bash
# Run from the project's folder
go run cmd/GoforPomodoroCheck/main.go
```

An output that tells that all is ok will look like

```
Go for Pomodoro FOSS -- sanity check.

- [✅] appsettings.toml file

- [✅] Database connected

- [✅] Telegram API connection
     Authorized on account <your bot name>


```

The bot can work also without a connected database, but in that case you will
obviously lose persistence of the data after application closing or PC
shut-down. It is up to you to decide whether that's ok or not for your bot
instance.

Also, obviously, the bot **cannot** work without a valid API key or verified
Telegram API connection.

### Running the application from source

As in Go it is very simple to compile and run projects, you just need to
perform this command.

```bash
# Run from the project's folder
go run cmd/GoforPomodoroBot/main.go
```

### Building (and running) the application from source

The same applies to a build that will leave you an executable file. 

```bash
# Run from the project's folder
go build cmd/GoforPomodoroBot/main.go

# To execute
./main
```